package server

import (
	"fmt"
	"sync"
)

type Status string

/*const (
	ClientStatusNew        = Status("new")
	ClientStatusConnecting = Status("connecting")
	ClientStatusConnected  = Status("connected")
	ClientStatusExiting    = Status("exiting")
	ClientStatusUnknown    = Status("unknown")
)*/

//ClientHandler is in charge of Handle Client Lifecycle
type ClientHandler interface {
	Accept(PeerClient) error
	Remove(PeerClient) error
	Notify(Node, error) //Used to get notifications of Client conn failures
	Clients() map[string]PeerClient
}

type defaultClientHandler struct {
	watcher Watcher
	clients map[string]PeerClient
	mutex   sync.Mutex
}

func DefaultClientHandler() *defaultClientHandler {
	//watcher Watcher
	return &defaultClientHandler{
		clients: make(map[string]PeerClient),
	}
}

func (h *defaultClientHandler) Accept(p PeerClient) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	node := p.Node()
	if _, ok := h.clients[node.String()]; ok {
		return fmt.Errorf("Client Already registered")
	}

	h.clients[node.String()] = p

	return nil
}

func (h *defaultClientHandler) Notify(n Node, err error) {

}

func (h *defaultClientHandler) Remove(p PeerClient) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	node := p.Node()
	if _, ok := h.clients[node.String()]; !ok {
		return fmt.Errorf("Client Not found")
	}

	delete(h.clients, node.String())

	return nil
}

func (h *defaultClientHandler) Clients() map[string]PeerClient {
	return h.clients
}

func (h *defaultClientHandler) check() {
	for k, v := range h.clients {
		fmt.Println("Clients: id ", k, " v ", v)
	}
}
