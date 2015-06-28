package router

import (
	"sync"

	"log"

	"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/node"
	"github.com/marcosQuesada/mesh/peer"
)

type Router interface {
	Accept(*peer.Peer)
	Handle(peer.NodePeer, message.Message) message.Message
	Route(message.Message)
	RegisterHandler(message.MsgType, Handler)
	Exit()

	InitDialClient(destination node.Node) //HEre??? NO
}

//Handler represent a method to be invoked with message
//error will be returned on unexpected handling
type Handler func(peer.NodePeer, message.Message) (message.Message, error)

type defaultRouter struct {
	handlers map[message.MsgType]Handler
	mutex    sync.Mutex
	exit     chan bool
	from     node.Node
}

func New(n node.Node) *defaultRouter {
	return &defaultRouter{
		handlers: make(map[message.MsgType]Handler),
		exit:     make(chan bool),
		from:     n,
	}
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

func (r *defaultRouter) Accept(c *peer.Peer) {
	go func() {
		for {
			select {
			case msg, open := <-c.ReceiveChan():
				if !open {
					//log.Println("Closed receiveChan, exit")
					return
				}
				response := r.Handle(c, msg)
				if response != nil {
					c.Commit(response)
					if response.MessageType() == message.ABORT {
						c.Exit()
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

		//To handle a correct response will need casting!
		//return &message.Error{}
		return nil
	}

	result, err := fn(c, msg)
	if err != nil {
		log.Fatalf("Handler error on message type %s error %s", msg.MessageType(), err)

		//To handle a correct response will need casting!
		return &message.Error{}
	}

	return result
}

func (r *defaultRouter) Route(message.Message) {

}

func (r *defaultRouter) Exit() {
	close(r.exit)
}

func (r *defaultRouter) InitDialClient(destination node.Node) {
	log.Println("Starting Dial Client from Node ", r.from.String(), "destination: ", destination.String())
	//Blocking call, wait until connection success
	p := peer.NewDialer(r.from, destination)
	go p.Run()

	//Say Hello and wait response
	p.SayHello()

	r.Accept(p)
}
