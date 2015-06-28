package router

import (
	"sync"

	"log"

	"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/peer"
)

type Router interface {
	Accept(*peer.Peer)
	Handle(message.Message) message.Message
	Route(message.Message)
	RegisterHandler(message.MsgType, Handler)
	Exit()
}

//Handler represent a method to be invoked with message
//error will be returned on unexpected handling
type Handler func(message.Message) (message.Message, error)

type defaultRouter struct {
	handlers map[message.MsgType]Handler
	mutex    sync.Mutex
	exit     chan bool
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

func (r *defaultRouter) Handle(msg message.Message) message.Message {
	fn, ok := r.handlers[msg.MessageType()]
	if !ok {
		log.Fatalf("Handler %s not found", msg.MessageType())

		//To handle a correct response will need casting!
		return &message.Error{}
	}

	result, err := fn(msg)
	if err != nil {
		log.Fatalf("Handler error on message type %s error %s", msg.MessageType(), err)

		//To handle a correct response will need casting!
		return &message.Error{}
	}

	return result
}

func (r *defaultRouter) Accept(c *peer.Peer) {
	go func() {
		for {
			select {
			case msg, open := <-c.ReceiveChan():
				if !open {
					log.Println("Closed receiveChan, exit")
					return
				}
				response := r.Handle(msg)
				c.Send(response)

			case <-r.exit:
				return
			}
		}
	}()
}

func (r *defaultRouter) Route(message.Message) {

}

func (r *defaultRouter) Exit() {
	close(r.exit)
}
