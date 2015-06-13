package peer_handler

import (
	"fmt"
	"github.com/marcosQuesada/mesh/dispatcher"
	"github.com/marcosQuesada/mesh/message"
	n "github.com/marcosQuesada/mesh/node"
	"github.com/marcosQuesada/mesh/peer"
	"github.com/marcosQuesada/mesh/watch"
	"log"
	"sync"
	"time"
)

//PeerHandler is in charge of Handle Client Lifecycle
//Sends pings on ticker to check remote state
type PeerHandler interface {
	Handle(peer.NodePeer) message.Status
	Route(message.Message)
	Events() chan dispatcher.Event
	AggregatedChan() chan message.Message
	Len() int
}

type defaultPeerHandler struct {
	watcher   watch.Watcher
	peers     map[string]peer.NodePeer
	mutex     sync.Mutex
	from      n.Node
	eventChan chan dispatcher.Event
	peerChan  chan message.Message
}

func DefaultPeerHandler(node n.Node) *defaultPeerHandler {
	return &defaultPeerHandler{
		watcher:   watch.New(),
		peers:     make(map[string]peer.NodePeer),
		from:      node,
		eventChan: make(chan dispatcher.Event, 0),
		peerChan:  make(chan message.Message, 0),
	}
}

func (d *defaultPeerHandler) Handle(c peer.NodePeer) (response message.Status) {
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

				d.eventChan <- &peer.OnPeerAbortedEvent{
					Node:  msg.(*message.Hello).From,
					Event: peer.PeerStatusError,
				}
				response = peer.PeerStatusAbort

				return
			}
			c.Send(&message.Welcome{Id: msg.(*message.Hello).Id, From: d.from, Details: map[string]interface{}{"foo_bar": 1231}})

			//d.watcher.Watch(c)
			d.eventChan <- &peer.OnPeerConnectedEvent{
				Node:  msg.(*message.Hello).From,
				Event: peer.PeerStatusConnected,
				Peer:  c,
			}
			response = peer.PeerStatusConnected

		case *message.Welcome:
			err := d.accept(c)
			if err != nil {
				d.eventChan <- &peer.OnPeerErroredEvent{
					Node:  c.Node(),
					Event: peer.PeerStatusError,
				}

				response = peer.PeerStatusError
				return
			} else {
				d.eventChan <- &peer.OnPeerConnectedEvent{
					Node:  c.Node(),
					Event: peer.PeerStatusConnected,
					Peer:  c,
				}
				response = peer.PeerStatusConnected
			}
		case *message.Abort:
			d.eventChan <- &peer.OnPeerAbortedEvent{
				Node:  c.Node(),
				Event: peer.PeerStatusAbort,
			}
			response = peer.PeerStatusAbort
		default:
			log.Println("Unexpected type On response ")
		}
	}

	return
}

func (h *defaultPeerHandler) Events() chan dispatcher.Event {
	return h.eventChan
}

func (h *defaultPeerHandler) Route(m message.Message) {

}

func (h *defaultPeerHandler) AggregatedChan() chan message.Message {
	return h.peerChan
}

func (h *defaultPeerHandler) Len() int {
	return len(h.peers)
}

func (h *defaultPeerHandler) accept(p peer.NodePeer) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	node := p.Node()
	if _, ok := h.peers[node.String()]; ok {
		return fmt.Errorf("Peer: %s Already registered", node.String())
	}
	h.peers[node.String()] = p
	//Agregate receiving Chann
	h.aggregate(p.ReceiveChan())

	return nil
}

func (h *defaultPeerHandler) remove(p peer.NodePeer) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	node := p.Node()
	if _, ok := h.peers[node.String()]; !ok {
		return fmt.Errorf("Peer Not found")
	}

	delete(h.peers, node.String())
	//@TODO: Close aggregated channel

	return nil
}

func (h *defaultPeerHandler) aggregate(c chan message.Message) {
	go func() {
		for {
			select {
			case m, open := <-c:
				if !open {
					return
				}
				h.peerChan <- m
			}
		}
	}()
}
