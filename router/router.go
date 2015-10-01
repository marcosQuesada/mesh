package router

import (
	"fmt"
	"log"
	"reflect"
	"sync"

	"github.com/marcosQuesada/mesh/dispatcher"
	"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/node"
	"github.com/marcosQuesada/mesh/peer"
	"github.com/marcosQuesada/mesh/router/handler"
	"github.com/marcosQuesada/mesh/watch"
)

type Router interface {
	Accept(*peer.Peer)
	Handle(peer.NodePeer, message.Message) message.Message
	RegisterHandlersFromInstance(handler.MessageHandler)
	Events() chan dispatcher.Event
	AggregateChan(chan handler.Request)
	Exit()

	InitDialClient(destination node.Node) //HEre??? NO
}

type defaultRouter struct {
	from            node.Node
	peers           map[string]peer.NodePeer
	peerIDs         map[peer.ID]bool
	eventChan       chan dispatcher.Event
	exit            chan bool
	watcher         watch.Watcher
	handlers        map[message.MsgType]handler.Handler
	notifiers       map[message.MsgType]bool
	mutex           sync.Mutex
	requestListener *watch.RequestListener
}

func New(n node.Node) *defaultRouter {
	evChan := make(chan dispatcher.Event, 10)
	reqList := watch.NewRequestListener()
	w := watch.New(reqList, evChan, 10)

	r := &defaultRouter{
		from:            n,
		peers:           make(map[string]peer.NodePeer),
		peerIDs:         make(map[peer.ID]bool, 0),
		eventChan:       evChan,
		handlers:        make(map[message.MsgType]handler.Handler),
		notifiers:       make(map[message.MsgType]bool),
		exit:            make(chan bool),
		watcher:         w,
		requestListener: reqList,
	}

	//RegisterHandlers & Notifiers from watcher
	r.RegisterHandlersFromInstance(w)

	r.registerHandlers(r.Handlers())
	r.registerNotifiers(r.Notifiers())

	return r
}

func (r *defaultRouter) RegisterHandlersFromInstance(h handler.MessageHandler) {
	r.registerHandlers(h.Handlers())
	n, ok := h.(handler.NotifyHandler)
	if ok {
		r.registerNotifiers(n.Notifiers())
	}
}

func (r *defaultRouter) registerHandlers(handlers map[message.MsgType]handler.Handler) {
	for msg, h := range handlers {
		r.registerHandler(msg, h)
	}
}

func (r *defaultRouter) registerHandler(msgType message.MsgType, handler handler.Handler) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, ok := r.handlers[msgType]; ok {
		log.Fatal("Handler already registered")
		return
	}

	r.handlers[msgType] = handler
}

func (r *defaultRouter) registerNotifiers(notifiers map[message.MsgType]bool) {
	for msg, s := range notifiers {
		r.registerNotifier(msg, s)
	}
}

func (r *defaultRouter) registerNotifier(msgType message.MsgType, st bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, ok := r.notifiers[msgType]; ok {
		log.Fatal("Notifier already registered")
		return
	}

	r.notifiers[msgType] = st
}

func (r *defaultRouter) AggregateChan(ch chan handler.Request) {
	for {
		select {
		case msg, open := <-ch:
			if !open {
				return
			}
			go func(h handler.Request){
				response, err := r.route(h.Msg)
				if err != nil {
					log.Println("Forwarding Msg from aggregateChan ERROR", err, h)
					h.ResponseChan <- message.Error{Id: h.Msg.ID()}

					return
				}
				h.ResponseChan <- response
			}(msg)

		case <-r.exit:
			log.Println("Exit Aggregate Loop")
			return
		}
	}
}

func (r *defaultRouter) route(msg message.Message) (message.Message, error) {
	from := msg.Destination()
	peer, ok := r.peers[from.String()]
	if !ok {
		return nil, fmt.Errorf("Router error: Destination Peer Not found")
	}

	peer.Commit(msg)
	//Wait response
	response := r.requestListener.Transaction(msg.ID())

	return response, nil
}

func (r *defaultRouter) Accept(c *peer.Peer) {
	go func() {
		for {
			select {
			case <-r.exit:
				log.Println("Exit", c.From(), c.Mode())
				return
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

				//@TODO: REMOVE IT, JUST DEVELOPMENT
				cmdData := ""
				if cmd, ok := msg.(*message.Command); ok {
					cmdData = reflect.TypeOf(cmd.Command).String()
				}
				log.Println("----------------RCV ", reflect.TypeOf(msg).String(), msg.ID(), msg.Origin(), cmdData)

				requestID := msg.ID()
				v, ok := r.notifiers[msg.MessageType()]
				if !ok {
					log.Fatal("Notifier not found ", msg.MessageType())
				}
				if v {
					r.requestListener.Notify(msg, requestID)
				}

				response := r.Handle(c, msg)
				if response != nil {
					c.Commit(response)

					if response.MessageType() == message.ABORT {
						c.Exit()
						return
					}

					if response.MessageType() == message.WELCOME {
						go r.requestListener.Transaction(requestID)
					}
				}
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
	p, requestID := peer.InitDialClient(r.from, destination)
	go r.requestListener.Transaction(requestID)
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
