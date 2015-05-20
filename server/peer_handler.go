package server

import (
	"fmt"
	"net"
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
	Accept(Peer) error
	Remove(Peer) error
	Notify(ID, error) //Used to get notifications of peer conn failures
	Peers() map[ID]Peer
}

type defaultPeerHandler struct {
	watcher Watcher
	peers   map[ID]Peer
	remotes map[net.Addr]ID
	mutex   sync.Mutex
}

func DefaultPeerHandler() *defaultPeerHandler {
	//watcher Watcher
	return &defaultPeerHandler{
		peers: make(map[ID]Peer),
	}
}

func (h *defaultPeerHandler) Accept(p Peer) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if _, ok := h.peers[p.Id()]; ok {
		return fmt.Errorf("Peer Already registered")
	}

	h.peers[p.Id()] = p
	h.remotes[p.Remote()] = p.Id()
	fmt.Println("Accepted Peer", p.Id())
	n := p.(*SocketPeer)
	fmt.Println("Accepted Peer", n.Conn.LocalAddr())

	return nil
}

func (h *defaultPeerHandler) Notify(id ID, err error) {

}

func (h *defaultPeerHandler) Remove(p Peer) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if _, ok := h.peers[p.Id()]; !ok {
		return fmt.Errorf("Peer Not found")
	}

	delete(h.peers, p.Id())
	delete(h.remotes, p.Remote())

	fmt.Println("Removed Peer", p.Id())

	return nil
}

func (h *defaultPeerHandler) Peers() map[ID]Peer {
	return h.peers
}

func (h *defaultPeerHandler) check() {
	for k, v := range h.peers {
		fmt.Println("Peers: id ", k, " v ", v)
	}
}
