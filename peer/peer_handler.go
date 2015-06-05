package peer

import (
	"fmt"
	"github.com/marcosQuesada/mesh/message"
	n "github.com/marcosQuesada/mesh/node"
	"github.com/marcosQuesada/mesh/watch"
	"sync"
)

//PeerHandler is in charge of Handle Client Lifecycle
//Sends pings on ticker to check remote state
type PeerHandler interface {
	Accept(NodePeer) error
	Remove(NodePeer) error
	Notify(n.Node, error) //Used to get notifications of Client conn failures
	Peers() map[string]NodePeer
	Len() int
	Route(message.Message)
}

type defaultPeerHandler struct {
	watcher watch.Watcher
	peers   map[string]NodePeer
	mutex   sync.Mutex
}

func DefaultPeerHandler() *defaultPeerHandler {
	return &defaultPeerHandler{
		watcher: watch.New(),
		peers:   make(map[string]NodePeer),
	}
}

func (h *defaultPeerHandler) Accept(p NodePeer) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	node := p.Node()
	if _, ok := h.peers[node.String()]; ok {
		return fmt.Errorf("Peer: %s Already registered", node.String())
	}
	h.peers[node.String()] = p
	fmt.Println("XX Accepted Peer type:", p.Mode(), " from: ", p.Node())
	return nil
}

func (h *defaultPeerHandler) Notify(n n.Node, err error) {

}

func (h *defaultPeerHandler) Remove(p NodePeer) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	node := p.Node()
	if _, ok := h.peers[node.String()]; !ok {
		return fmt.Errorf("Peer Not found")
	}

	delete(h.peers, node.String())

	return nil
}

func (h *defaultPeerHandler) Peers() map[string]NodePeer {
	return h.peers
}

func (h *defaultPeerHandler) Route(m message.Message) {

}

func (h *defaultPeerHandler) check() {
	for k, v := range h.peers {
		fmt.Println("peers: id ", k, " v ", v)
	}
}

func (h *defaultPeerHandler) Len() int {
	return len(h.peers)
}
