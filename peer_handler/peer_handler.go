package peer_handler

import (
	//"fmt"
	"fmt"
	"log"
	"sync"

	"github.com/marcosQuesada/mesh/dispatcher"
	"github.com/marcosQuesada/mesh/message"
	n "github.com/marcosQuesada/mesh/node"
	"github.com/marcosQuesada/mesh/peer"
	"github.com/marcosQuesada/mesh/watch"
)

//PeerHandler is in charge of Handle Client Lifecycle
//Sends pings on ticker to check remote state
type PeerHandler interface {
	HandleHello(c peer.NodePeer, msg message.Message) (message.Message, error)
	HandleWelcome(c peer.NodePeer, msg message.Message) (message.Message, error)
	HandleAbort(c peer.NodePeer, msg message.Message) (message.Message, error)
	HandleDone(c peer.NodePeer, msg message.Message) (message.Message, error)
	HandleError(c peer.NodePeer, msg message.Message) (message.Message, error)

	Remove(peer.NodePeer) error
	Route(message.Message)
	Events() chan dispatcher.Event
}

type defaultPeerHandler struct {
	watcher   watch.Watcher
	peers     map[string]peer.NodePeer
	from      n.Node
	eventChan chan dispatcher.Event
	mutex     sync.Mutex
}

func DefaultPeerHandler(node n.Node) *defaultPeerHandler {
	evCh := make(chan dispatcher.Event, 10)
	return &defaultPeerHandler{
		watcher:   watch.New(evCh, 2),
		peers:     make(map[string]peer.NodePeer),
		from:      node,
		eventChan: evCh,
	}
}

func (d *defaultPeerHandler) HandleHello(c peer.NodePeer, msg message.Message) (message.Message, error) {
	log.Println("Peer Handler Hello", msg)
	c.Identify(msg.(*message.Hello).From)
	err := d.accept(c)
	if err != nil {
		c.Exit()

		d.eventChan <- &peer.OnPeerAbortedEvent{
			Node:  msg.(*message.Hello).From,
			Event: peer.PeerStatusError,
		}

		return &message.Abort{Id: msg.(*message.Hello).Id, From: d.from}, nil
	}

	//go d.watcher.Watch(c)
	d.eventChan <- &peer.OnPeerConnectedEvent{
		Node:  msg.(*message.Hello).From,
		Event: peer.PeerStatusConnected,
		Peer:  c,
	}

	return &message.Welcome{Id: msg.(*message.Hello).Id, From: d.from}, nil
}

func (d *defaultPeerHandler) HandleWelcome(c peer.NodePeer, msg message.Message) (message.Message, error) {
	log.Println("Peer Handler Welcome", msg)
	err := d.accept(c)
	if err != nil {
		d.eventChan <- &peer.OnPeerErroredEvent{
			Node:  c.Node(),
			Event: peer.PeerStatusError,
			Error: err,
		}

		return &message.Error{Id: msg.(*message.Welcome).Id, From: d.from}, err
	} else {
		//go d.watcher.Watch(c)
		d.eventChan <- &peer.OnPeerConnectedEvent{
			Node:  c.Node(),
			Event: peer.PeerStatusConnected,
			Peer:  c,
		}
		return &message.Done{Id: msg.(*message.Welcome).Id, From: d.from}, nil
	}
}

func (d *defaultPeerHandler) HandleAbort(c peer.NodePeer, msg message.Message) (message.Message, error) {
	log.Println("Peer Handler Abort", msg)
	d.eventChan <- &peer.OnPeerAbortedEvent{
		Node:  c.Node(),
		Event: peer.PeerStatusAbort,
	}

	c.Exit()
	return &message.Done{Id: msg.(*message.Abort).Id, From: d.from}, nil
}

func (d *defaultPeerHandler) HandleDone(c peer.NodePeer, msg message.Message) (message.Message, error) {
	log.Println("Peer Handler Done", msg)
	return nil, nil
}

func (d *defaultPeerHandler) HandleError(c peer.NodePeer, msg message.Message) (message.Message, error) {
	log.Println("Peer Handler Error", msg)
	return nil, nil
}

func (h *defaultPeerHandler) Events() chan dispatcher.Event {
	return h.eventChan
}

func (h *defaultPeerHandler) Route(m message.Message) {

}

func (h *defaultPeerHandler) accept(p peer.NodePeer) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	node := p.Node()
	if _, ok := h.peers[node.String()]; ok {
		return fmt.Errorf("Peer: %s Already registered", node.String())
	}
	h.peers[node.String()] = p

	return nil
}

func (h *defaultPeerHandler) Remove(p peer.NodePeer) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	node := p.Node()
	if _, ok := h.peers[node.String()]; !ok {
		return fmt.Errorf("Peer Not found")
	}
	fmt.Println("PeerHandler Removed Peer ", node.String())
	delete(h.peers, node.String())
	//@TODO: Close aggregated channel

	return nil
}
