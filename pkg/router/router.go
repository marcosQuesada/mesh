package router

import (
	"fmt"
	"log"
	"reflect"
	"sync"

	"github.com/marcosQuesada/mesh/pkg/cli"
	"github.com/marcosQuesada/mesh/pkg/dispatcher"
	"github.com/marcosQuesada/mesh/pkg/message"
	"github.com/marcosQuesada/mesh/pkg/node"
	"github.com/marcosQuesada/mesh/pkg/peer"
	"github.com/marcosQuesada/mesh/pkg/router/handler"
	"github.com/marcosQuesada/mesh/pkg/router/request"
	"github.com/marcosQuesada/mesh/pkg/watch"
)

type Router interface {
	Accept(*peer.Peer)
	RegisterHandlersFromInstance(handler.MessageHandler)
	AggregateChan(chan handler.Request)
	CliHandlers() map[string]cli.Definition
	RequestListener() *request.RequestListener
	Exit()
}

type defaultRouter struct {
	from            *node.Node               //Origin node
	peers           map[string]peer.PeerNode //neighbours
	peerIDs         map[peer.ID]bool
	events          chan dispatcher.Event
	exit            chan bool
	watcher         watch.Watcher
	handlers        map[message.MsgType]handler.Handler
	notifiers       map[message.MsgType]bool
	transactionals  map[message.MsgType]bool
	mutex           *sync.Mutex
	requestListener *request.RequestListener
	dispatcher      dispatcher.Dispatcher
}

func New(n *node.Node, dispatcher dispatcher.Dispatcher) *defaultRouter {
	evChan := dispatcher.SndChan()
	reqList := request.NewRequestListener()
	w := watch.New(reqList, evChan, 10)

	r := &defaultRouter{
		from:            n,
		peers:           make(map[string]peer.PeerNode),
		peerIDs:         make(map[peer.ID]bool, 0),
		events:          evChan,
		exit:            make(chan bool),
		watcher:         w,
		handlers:        make(map[message.MsgType]handler.Handler),
		notifiers:       make(map[message.MsgType]bool),
		transactionals:  make(map[message.MsgType]bool),
		mutex:           &sync.Mutex{},
		requestListener: reqList,
		dispatcher:      dispatcher,
	}

	//RegisterHandlers & Notifiers from watcher
	r.RegisterHandlersFromInstance(w)

	//Register locals
	r.registerHandlers(r.Handlers())
	r.registerNotifiers(r.Notifiers())
	r.registerTransactionals(r.Transactions())

	return r
}

func (r *defaultRouter) RequestListener() *request.RequestListener {
	return r.requestListener
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
					r.peerDisconnected(c)
					return
				}

				r.logCommand(msg)

				// Notify waiting listener if required
				if v, ok := r.notifiers[msg.MessageType()]; ok && v {
					r.requestListener.Notify(msg, msg.ID())
				}

				response := r.handle(c, msg)
				if response != nil {
					c.Commit(response)

					// Start new request transaction if required
					if r.isTransactional(msg) {
						go r.requestListener.Register(response.ID()) //@TODO: RETHINK SYNC
					}

					// On abort disconnect acceptor
					if response.MessageType() == message.ABORT {
						c.Exit()
						return
					}
				}
			}
		}
	}()
}

func (r *defaultRouter) Exit() {
	close(r.exit)
}

func (r *defaultRouter) handle(c peer.PeerNode, msg message.Message) message.Message {
	fn, ok := r.handlers[msg.MessageType()]
	if !ok {
		log.Fatalf("Handler %v not found", msg.MessageType())

		return &message.Error{}
	}

	result, err := fn(c, msg)
	if err != nil {
		log.Fatalf("Handler error on message type %v error %s", msg.MessageType(), err)

		//To handle a correct response will need casting!
		return &message.Error{}
	}

	return result
}

func (r *defaultRouter) accept(p peer.PeerNode) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	var node = p.Node()
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
		return nil, fmt.Errorf("Router error: Destination Peer Not found %s", to.String())
	}

	peer.Commit(msg)

	//Open transaction if required and wait response
	if r.isTransactional(msg) {
		return r.requestListener.RegisterAndWait(msg.ID())
	}

	return nil, nil
}

func (r *defaultRouter) peerDisconnected(c *peer.Peer) {
	if _, ok := r.peerIDs[c.Id()]; ok {
		log.Println("Unregister Peer:", c.Node(), "mode:", c.Mode(), "id", c.Id())
		r.removePeer(c)

		r.dispatcher.Dispatch(&peer.OnPeerDisconnectedEvent{c.Node(), peer.PeerStatusDisconnected})
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

func (r *defaultRouter) isTransactional(msg message.Message) bool {
	v, ok := r.transactionals[msg.MessageType()]
	if !ok {
		log.Fatal("Transactioner not found ", msg.MessageType())
	}

	return v && ok
}

func (r *defaultRouter) registerTransactionals(transactionals map[message.MsgType]bool) {
	for msg, s := range transactionals {
		r.registerTransactional(msg, s)
	}
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

func (r *defaultRouter) peerExists(p peer.PeerNode) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	var node = p.Node()
	_, ok := r.peers[node.String()]

	return ok
}

func (r *defaultRouter) removePeer(p peer.PeerNode) error {
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

// CliHandlers exports command cli definitions
func (r *defaultRouter) CliHandlers() map[string]cli.Definition {
	return map[string]cli.Definition{}
}

func (r *defaultRouter) logCommand(msg message.Message) {
	cmdData := ""
	if cmd, ok := msg.(*message.Command); ok {
		cmdData = reflect.TypeOf(cmd.Command).String()
	}
	log.Println("---RCV ", reflect.TypeOf(msg).String(), msg.ID(), msg.Origin(), cmdData)
}
