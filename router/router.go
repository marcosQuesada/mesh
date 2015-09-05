package router

import (
	"fmt"
	"log"
	"sync"
	"reflect"

	"github.com/marcosQuesada/mesh/dispatcher"
	"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/node"
	"github.com/marcosQuesada/mesh/peer"
	"github.com/marcosQuesada/mesh/router/handler"
	"github.com/marcosQuesada/mesh/watch"
)

type Router interface {
	Route(message.Message) error
	Accept(*peer.Peer)
	Handle(peer.NodePeer, message.Message) message.Message
	RegisterHandlers(handlers map[message.MsgType]handler.Handler)
	RegisterHandler(message.MsgType, handler.Handler)
	Events() chan dispatcher.Event
	Exit()

	InitDialClient(destination node.Node) //HEre??? NO
}

type defaultRouter struct {
	handlers        map[message.MsgType]handler.Handler
	from            node.Node
	peers           map[string]peer.NodePeer
	peerIDs         map[peer.ID]bool
	eventChan       chan dispatcher.Event
	exit            chan bool
	watcher         watch.Watcher
	mutex           sync.Mutex
	requestListener *watch.RequestListener
}

func New(n node.Node) *defaultRouter {
	evChan := make(chan dispatcher.Event, 10)
	reqList := watch.NewRequestListener()
	w := watch.New(reqList, evChan, 2)

	r := &defaultRouter{
		handlers:        make(map[message.MsgType]handler.Handler),
		from:            n,
		peers:           make(map[string]peer.NodePeer),
		peerIDs:         make(map[peer.ID]bool, 0),
		eventChan:       evChan,
		exit:            make(chan bool),
		watcher:         w,
		requestListener: reqList,
	}

	r.RegisterHandlers(r.Handlers())
	r.RegisterHandlers(w.Handlers())

	return r
}

func (r *defaultRouter) RegisterHandlers(handlers map[message.MsgType]handler.Handler) {
	for msg, h := range handlers {
		fmt.Println("Registering ", msg)
		r.RegisterHandler(msg, h)
	}
}

func (r *defaultRouter) RegisterHandler(msgType message.MsgType, handler handler.Handler) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, ok := r.handlers[msgType]; ok {
		log.Fatal("Handler already registered")
		return
	}

	r.handlers[msgType] = handler
}

func (r *defaultRouter) Route(msg message.Message) error {
	from := msg.Destination()
	peer, ok := r.peers[from.String()]
	if !ok {
		return fmt.Errorf("Peer Not found")
	}

	peer.Commit(msg)

	return nil
}

func (r *defaultRouter) WaitResponse(c *peer.Peer, requestId message.ID) {
	r.requestListener.Transaction(requestId)
}

func (r *defaultRouter) Accept(c *peer.Peer) {
	go func() {
		for {
			select {
			case msg, open := <-c.ReceiveChan():
				if !open {
					if _, ok := r.peerIDs[c.Id()]; ok {
						log.Println("Unregister Peer:", c.Node(), "mode:", c.Mode(), "id", c.Id())
						r.remove(c)
						r.eventChan <- &peer.OnPeerDisconnectedEvent{c.From(), peer.PeerStatusDisconnected}
						go r.InitDialClient(c.Node())
					}

					return
				}
				log.Println("----------------RCV ", reflect.TypeOf(msg).String(), msg.ID(), msg.Origin())

				requestID := msg.ID()
				response := r.Handle(c, msg)
				if response != nil {
					c.Commit(response)

					//TODO: Solve this type handling mess
					if response.MessageType() == message.ABORT {
						c.Exit()
						return
					}

					if response.MessageType() == message.ACK {
						r.requestListener.Notify(msg, requestID)
						continue
					}

					if response.MessageType() != message.PONG {
						r.WaitResponse(c, requestID)
					}
				}

				if msg.MessageType() != message.HELLO && msg.MessageType() != message.PING  {
					r.requestListener.Notify(msg, requestID)
				}
			case <-r.exit:
				log.Println("Exit", c.From(), c.Mode())
				return
			}
		}
	}()
}

func (r *defaultRouter) Handle(c peer.NodePeer, msg message.Message) message.Message {
	fn, ok := r.handlers[msg.MessageType()]
	if !ok {
		log.Fatalf("Handler %s not found", msg.MessageType())

		return &message.Error{}
	}

	result, err := fn(c, msg)
	if err != nil {
		log.Fatalf("Handler error on message type %s error %s", msg.MessageType(), err)

		//To handle a correct response will need casting!
		return &message.Error{}
	}

	return result
}

func (r *defaultRouter) Exit() {
	close(r.exit)
}

func (r *defaultRouter) Events() chan dispatcher.Event {
	return r.eventChan
}

func (r *defaultRouter) InitDialClient(destination node.Node) {
	p, requestId := peer.InitDialClient(r.from, destination)
	r.WaitResponse(p, requestId)
	r.Accept(p)
}

func (r *defaultRouter) accept(p peer.NodePeer) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	var node node.Node = p.Node()
	if _, ok := r.peers[node.String()]; ok {
		return fmt.Errorf("Peer: %s Already registered", node.String())
	}

	r.peers[node.String()] = p
	r.peerIDs[p.Id()] = true

	return nil
}

func (r *defaultRouter) exists(p peer.NodePeer) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	var node node.Node = p.Node()
	_, ok := r.peers[node.String()]

	return ok
}

func (r *defaultRouter) remove(p peer.NodePeer) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	node := p.Node()
	if _, ok := r.peers[node.String()]; !ok {
		return fmt.Errorf("Peer Not found")
	}

	delete(r.peers, node.String())
	delete(r.peerIDs, p.Id())

	return nil
}
