package watch

//Takes care to regular pull peers to update its link state
// Adds a ticker on each peer added,
// Each watcher is a goroutine sending pooling messages to a channel
// Master goroutine forwards pool requests to peers and awaits answer

// Basic PING PONG Mechanism

import (
	"log"
	"sync"
	"time"

	"github.com/marcosQuesada/mesh/dispatcher"
	"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/peer"
)

type Watcher interface {
	Watch(peer.NodePeer)
	HandlePing(peer.NodePeer, message.Message) (message.Message, error)
	HandlePong(peer.NodePeer, message.Message) (message.Message, error)
}

type defaultWatcher struct {
	eventChan       chan dispatcher.Event
	exit            chan bool
	pingInterval    int
	mutex           sync.RWMutex
	index           map[string]*subject
	requestListener *requestListener
	wg              sync.WaitGroup
}

type subject struct {
	peer        peer.NodePeer
	ticker      *time.Ticker
	id          int
	Done        chan bool
	tickerReset chan bool
	mutex       sync.Mutex
}

func (s *subject) setId(id int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.id = id
}

func (s *subject) getId() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.id
}

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
}
func New(evCh chan dispatcher.Event, interval int) *defaultWatcher {
	rl := &requestListener{
		listeners: make(map[string]chan message.Message, 0),
		timeout:   time.Duration(2),
	}

	return &defaultWatcher{
		eventChan:       evCh,
		exit:            make(chan bool, 0),
		index:           make(map[string]*subject, 0),
		pingInterval:    interval,
		requestListener: rl,
	}
}

func (w *defaultWatcher) Watch(p peer.NodePeer) {
	defer log.Println("Watcher", p.Node(), "Exits")
	//add watcher to waitGroup
	log.Println("Started Watcher to", p.Node())
	w.wg.Add(1)
	defer w.wg.Done()

	//@TODO: Randomize this
	tickerReset := make(chan bool, 10)
	subjectDone := make(chan bool, 10)
	ticker := newTicker(w.pingInterval)
	s := &subject{
		peer:        p,
		ticker:      ticker,
		Done:        subjectDone,
		tickerReset: tickerReset,
	}
	defer close(s.Done)
	defer close(s.tickerReset)

	node := p.Node()
	w.mutex.Lock()
	w.index[node.String()] = s
	w.mutex.Unlock()

	for {
		select {
		case <-s.ticker.C:
			id := s.getId()
			p.Send(&message.Ping{Id: id, From: p.From(), To: node})
			log.Println("Sended PING", id, "to", node.String())
			requestId := w.requestListener.Id(p.Node(), id)
			w.requestListener.register(requestId)
			log.Println("Waiting ", requestId)

			msg, err := w.requestListener.wait(requestId)
			if err != nil {
				log.Println("EEEEEEEEEEE Error Waiting requestListener ", requestId, err)
			} else {
				log.Println("XXX Message is ", msg)
			}

			id++
			s.setId(id)

		case <-s.Done:
			return
		}
	}

}

func (w *defaultWatcher) Exit() {
	close(w.exit)

	//stop all Subject watchers
	for s := range w.index {
		w.index[s].Done <- true
		log.Println("Stoping Watchers")
	}

	w.wg.Wait()
	log.Println("Exiting Done")
}

func (w *defaultWatcher) HandlePing(c peer.NodePeer, msg message.Message) (message.Message, error) {
	ping := msg.(*message.Ping)
	log.Println("Received Ping", ping.Id, "from: ", ping.From.String())

	return &message.Pong{Id: ping.Id, From: ping.To, To: ping.From}, nil
}

func (w *defaultWatcher) HandlePong(c peer.NodePeer, msg message.Message) (message.Message, error) {
	log.Println("Handle Pong ", c.Node(), msg)
	pong := msg.(*message.Pong)
	requestID := w.requestListener.Id(c.Node(), pong.Id)
	go w.requestListener.notify(msg, requestID)

	return nil, nil
}

func newTicker(pingInterval int) *time.Ticker {
	return time.NewTicker(time.Duration(pingInterval) * time.Second)
}
