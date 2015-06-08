package peer

import (
	"fmt"
	"github.com/marcosQuesada/mesh/message"
	n "github.com/marcosQuesada/mesh/node"
	"github.com/marcosQuesada/mesh/watch"
	"log"
	"sync"
	"time"
)

//PeerHandler is in charge of Handle Client Lifecycle
//Sends pings on ticker to check remote state
type PeerHandler interface {
	Handle(NodePeer) message.Status
	Route(message.Message)
	Events() chan MemberUpdate
	Len() int
}

type defaultPeerHandler struct {
	watcher   watch.Watcher
	peers     map[string]NodePeer
	mutex     sync.Mutex
	from      n.Node
	eventChan chan MemberUpdate
}
type MemberUpdate struct {
	Node  n.Node
	Event message.Status
	Peer  NodePeer
}

func DefaultPeerHandler(node n.Node) *defaultPeerHandler {
	return &defaultPeerHandler{
		watcher:   watch.New(),
		peers:     make(map[string]NodePeer),
		from:      node,
		eventChan: make(chan MemberUpdate, 0),
	}
}

func (d *defaultPeerHandler) Handle(c NodePeer) (response message.Status) {
	node := c.Node()
	select {
	case <-time.NewTimer(time.Second).C:
		log.Println("Client has not receive response, Timeout")
		return
	case msg := <-c.ReceiveChan():
		switch msg.(type) {
		case *message.Hello:
			c.Identify(msg.(*message.Hello).From)
			err := d.accept(c)
			if err != nil {
				c.Send(&message.Abort{Id: msg.(*message.Hello).Id, From: d.from, Details: map[string]interface{}{"foo_bar": 1231}})
				c.Exit()

				d.eventChan <- MemberUpdate{
					Node:  msg.(*message.Hello).From,
					Event: PeerStatusError,
				}

				response = PeerStatusAbort
				return
			}
			c.Send(&message.Welcome{Id: msg.(*message.Hello).Id, From: d.from, Details: map[string]interface{}{"foo_bar": 1231}})

			d.eventChan <- MemberUpdate{
				Node:  msg.(*message.Hello).From, //Sure??
				Event: PeerStatusConnected,
				Peer:  c,
			}
			response = PeerStatusConnected
			//log.Println("Client Achieved: ", msg.(*message.Hello).From)
		case *message.Welcome:
			//log.Println("Client has received Welcome from", node.String(), msg.(*message.Welcome))
			err := d.accept(c)
			if err != nil {
				//log.Println("Error Accepting Peer, Peer dies! ", err)
				d.eventChan <- MemberUpdate{
					Node:  node,
					Event: PeerStatusError,
				}

				response = PeerStatusError
				return
			} else {
				d.eventChan <- MemberUpdate{
					Node:  node,
					Event: PeerStatusConnected,
					Peer:  c,
				}
				response = PeerStatusConnected
				//log.Println("Client Achieved: ", node)
			}
		case *message.Abort:
			//log.Println("Response Abort ", msg.(*message.Abort), " remote node:", node.String())
			d.eventChan <- MemberUpdate{
				Node:  node,
				Event: PeerStatusAbort,
			}
			response = PeerStatusAbort
		default:
			log.Println("Unexpected type On response ")
		}
	}

	return
}

func (h *defaultPeerHandler) Events() chan MemberUpdate {
	return h.eventChan
}

func (h *defaultPeerHandler) Route(m message.Message) {

}

func (h *defaultPeerHandler) Len() int {
	return len(h.peers)
}

func (h *defaultPeerHandler) accept(p NodePeer) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	node := p.Node()
	if _, ok := h.peers[node.String()]; ok {
		return fmt.Errorf("Peer: %s Already registered", node.String())
	}
	h.peers[node.String()] = p

	return nil
}

func (h *defaultPeerHandler) remove(p NodePeer) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	node := p.Node()
	if _, ok := h.peers[node.String()]; !ok {
		return fmt.Errorf("Peer Not found")
	}

	delete(h.peers, node.String())

	return nil
}
