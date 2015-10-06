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
	transactionals map[message.MsgType]bool
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
		transactionals:  make(map[message.MsgType]bool),
		exit:            make(chan bool),
		watcher:         w,
		requestListener: reqList,
	}

	//RegisterHandlers & Notifiers from watcher
	r.RegisterHandlersFromInstance(w)

	//Register locals
	r.registerHandlers(r.Handlers())
	r.registerNotifiers(r.Notifiers())
	r.registerTransactionals(r.Transactions())

	return r
}

func (r *defaultRouter) RegisterHandlersFromInstance(h handler.MessageHandler) {
	r.registerHandlers(h.Handlers())
	n, ok := h.(handler.NotifyHandler)
	if ok {
		r.registerNotifiers(n.Notifiers())
	}
	t, ok := h.(handler.TransactionHandler)
	if ok {
		r.registerTransactionals(t.Transactions())
	}
}

func (r *defaultRouter) AggregateChan(ch chan handler.Request) {
	for {
		select {
		case msg, open := <-ch:
			if !open {
				return
			}
			go func(h handler.Request) {
				response, err := r.route(h.Msg.(message.Message))
				if err != nil {
					log.Println("Forwarding Msg from aggregateChan ERROR", err, h)
					h.ResponseChan <- message.Error{Id: h.Msg.ID(), Err: err}

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

func (r *defaultRouter) Accept(c *peer.Peer) {
	go func() {
		for {
			select {
			case <-r.exit:
				log.Println("Exit", c.From(), c.Mode())
				return
			case msg, open := <-c.ReceiveChan():
				if !open {
					r.startPeerRecovering(c)
					return
				}

				r.logCommand(msg)
				r.notifyTransaction(msg)
				response := r.Handle(c, msg)
				if response != nil {
					c.Commit(response)
					r.startTransaction(response)

					if response.MessageType() == message.ABORT {
						c.Exit()
						return
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

func (r *defaultRouter) Events() chan dispatcher.Event {
	return r.eventChan
}

func (r *defaultRouter) InitDialClient(destination node.Node) {
	p, requestID := peer.InitDialClient(r.from, destination)
	go r.requestListener.Transaction(requestID)
	r.Accept(p)
}

func (r *defaultRouter) Exit() {
	close(r.exit)
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

func (r *defaultRouter) route(msg message.Message) (message.Message, error) {
	to := msg.Destination()
	peer, ok := r.peers[to.String()]
	if !ok {
		return nil, fmt.Errorf("Router error: Destination Peer Not found ", to.String())
	}

	peer.Commit(msg)

	//Open transaction iif required and wait response
	v, ok := r.transactionals[msg.MessageType()]
	if !ok {
		log.Fatal("Transactioner not found ", msg.MessageType())
	}
	if v {
		response := r.requestListener.Transaction(msg.ID())

		return response, nil
	}

	return nil, nil
}

func (r *defaultRouter) startPeerRecovering(c *peer.Peer) {
	if _, ok := r.peerIDs[c.Id()]; ok {
		log.Println("Unregister Peer:", c.Node(), "mode:", c.Mode(), "id", c.Id())
		r.removePeer(c)
		r.eventChan <- &peer.OnPeerDisconnectedEvent{c.From(), peer.PeerStatusDisconnected}
		go r.InitDialClient(c.Node())
	}
}

func (r *defaultRouter) logCommand(msg message.Message) {
	cmdData := ""
	if cmd, ok := msg.(*message.Command); ok {
		cmdData = reflect.TypeOf(cmd.Command).String()
	}
	log.Println("---RCV ", reflect.TypeOf(msg).String(), msg.ID(), msg.Origin(), cmdData)
}

func (r *defaultRouter) notifyTransaction(msg message.Message) {
	v, ok := r.notifiers[msg.MessageType()]
	if !ok {
		log.Fatal("Notifier not found ", msg.MessageType())
	}
	if v {
		r.requestListener.Notify(msg, msg.ID())
	}
}

func (r *defaultRouter) startTransaction(msg message.Message) {
	v, ok := r.transactionals[msg.MessageType()]
	if !ok {
		log.Fatal("Transactioner not found ", msg.MessageType())
	}

	if v {
		go r.requestListener.Transaction(msg.ID())
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

func (r *defaultRouter) registerTransactionals(transactionals map[message.MsgType]bool) {
	for msg, s := range transactionals {
		r.registerTransactional(msg, s)
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

func (r *defaultRouter) registerTransactional(msgType message.MsgType, st bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, ok := r.transactionals[msgType]; ok {
		log.Fatal("Transactioner already registered")
		return
	}

	r.transactionals[msgType] = st
}

func (r *defaultRouter) existPeer(p peer.NodePeer) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	var node node.Node = p.Node()
	_, ok := r.peers[node.String()]

	return ok
}

func (r *defaultRouter) removePeer(p peer.NodePeer) error {
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
