package server

import (
	"fmt"
	"sync"
)

type Status string

const (
	PeerStatusNew        = Status("new")
	PeerStatusConnecting = Status("connecting")
	PeerStatusConnected  = Status("connected")
	PeerStatusExiting    = Status("exiting")
	PeerStatusUnknown    = Status("unknown")
)

//PeerHandler is in charge of Handle Peer Lifecycle
type PeerHandler interface {
	Accept(*SocketPeer) error
	Remove(Peer)
	Notify(ID, error) //Used to get notifications of peer conn failures
}

type defaultPeerHandler struct {
	watcher Watcher
	peers   map[ID]Peer
	//What about nodes??
	mutex sync.Mutex
}

func (h *defaultPeerHandler) Accept(p *SocketPeer) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.peers[p.Id()] = p

	return nil
}

func (h *defaultPeerHandler) Notify(id ID, err error) {

}

func (h *defaultPeerHandler) Remove(p Peer) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	delete(h.peers, p.Id())
}

func (h *defaultPeerHandler) check() {
	for k, v := range h.peers {
		fmt.Println("Peers: id ", k, " v ", v)
	}
}
