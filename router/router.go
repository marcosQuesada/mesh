package router

import (
	"fmt"
	"sync"

	"log"

	"github.com/marcosQuesada/mesh/dispatcher"
	"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/node"
	"github.com/marcosQuesada/mesh/peer"
	"github.com/marcosQuesada/mesh/watch"
)

type Router interface {
	Route(message.Message) error
	Accept(*peer.Peer)
	Handle(peer.NodePeer, message.Message) message.Message
	RegisterHandler(message.MsgType, Handler)
	Events() chan dispatcher.Event
	Exit()

	InitDialClient(destination node.Node) //HEre??? NO
}

//Handler represent a method to be invoked with message
//error will be returned on unexpected handling
type Handler func(peer.NodePeer, message.Message) (message.Message, error)

type defaultRouter struct {
	handlers  map[message.MsgType]Handler
	from      node.Node
	peers     map[string]peer.NodePeer
	peerIDs   map[peer.ID]bool
	eventChan chan dispatcher.Event
	exit      chan bool
	watcher   watch.Watcher
	mutex     sync.Mutex
}

func New(n node.Node) *defaultRouter {
	evChan := make(chan dispatcher.Event, 10)
	w := watch.New(evChan, 2)
	r := &defaultRouter{
		from:      n,
		peers:     make(map[string]peer.NodePeer),
		peerIDs:   make(map[peer.ID]bool, 0),
		eventChan: evChan,
		exit:      make(chan bool),
		watcher:   w,
	}
	//initialize local handlers
	r.handlers = map[message.MsgType]Handler{
		message.HELLO:   r.HandleHello,
		message.WELCOME: r.HandleWelcome,
		message.ABORT:   r.HandleAbort,
		message.ERROR:   r.HandleError,
		message.PING:    w.HandlePing,
		message.PONG:    w.HandlePong,
	}

	return r
}

func (r *defaultRouter) RegisterHandler(msgType message.MsgType, handler Handler) {
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

				response := r.Handle(c, msg)
				if response != nil {
					c.Commit(response)
					if response.MessageType() == message.ABORT {
						c.Exit()
						return
					}
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
	//Blocking call, wait until connection success
	p := peer.NewDialer(r.from, destination)
	go p.Run()
	log.Println("Connected Dial Client from Node ", r.from.String(), "destination: ", destination.String(), p.Id())

	//Say Hello and wait response
	p.SayHello()
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
