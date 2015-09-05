package watch

//Takes care to regular pull peers to update its link state
// Adds a ticker on each peer added,
// Each watcher is a goroutine sending pooling messages to a channel
// Master goroutine forwards pool requests to peers and awaits answer

// Basic PING PONG Mechanism

import (
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/marcosQuesada/mesh/dispatcher"
	"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/peer"
	//"github.com/nu7hatch/gouuid"
)

/*type Watcher interface {
	Watch(peer.NodePeer)
	HandlePing(peer.NodePeer, message.Message) (message.Message, error)
	HandlePong(peer.NodePeer, message.Message) (message.Message, error)
}*/

type defaultWatcher struct {
	eventChan       chan dispatcher.Event
	exit            chan bool
	pingInterval    int
	mutex           sync.RWMutex
	index           map[string]*subject
	requestListener *RequestListener
	wg              sync.WaitGroup
}

type subject struct {
	peer   peer.NodePeer
	ticker *time.Ticker
	id     message.ID
	Done   chan bool
	mutex  sync.Mutex
}

func (s *subject) incId() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.id = message.NewId()
}

func (s *subject) getId() message.ID {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.id
}

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	rand.Seed(time.Now().UTC().UnixNano())
}

func New(evCh chan dispatcher.Event, interval int) *defaultWatcher {
	return &defaultWatcher{
		eventChan:       evCh,
		exit:            make(chan bool, 0),
		index:           make(map[string]*subject, 0),
		pingInterval:    interval * 1000,
		requestListener: NewRequestListener(),
	}
}

func (w *defaultWatcher) Watch(p peer.NodePeer) {
	defer log.Println("XXX Watcher", p.Node(), "Exits")

	//add watcher to waitGroup
	w.wg.Add(1)
	defer w.wg.Done()

	//@TODO: Randomize this
	subjectDone := make(chan bool, 10)
	s := &subject{
		peer:   p,
		ticker: w.newTicker(),
		Done:   subjectDone,
	}
	defer close(s.Done)

	node := p.Node()
	w.mutex.Lock()
	w.index[node.String()] = s
	w.mutex.Unlock()

	for {
		select {
		case <-s.ticker.C:
			p.Commit(&message.Ping{Id: s.getId(), From: p.From(), To: node})
			s.ticker.Stop()

			requestId := s.getId()
			log.Println("Registering PONG Watcher", requestId)
			w.requestListener.Register(requestId)
			log.Println("PING", s.getId(), "to", node.String(), "Waiting ", requestId)

			msg, err := w.requestListener.Wait(requestId)
			if err != nil {
				log.Println("RequestListener ", requestId, err)

				return
			}
			if msg.MessageType() != message.PONG {
				log.Println("Error Unexpected Received type, expected PONG ", msg.MessageType(), "RequestListener ", requestId, err)
				//@TODO: used to check development stability
				panic(err)
			}

			log.Println("PING PONG OK", s.getId(), "to", node.String(), "Waiting ", requestId)

			s.incId()
			s.ticker = w.newTicker()
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

func (w *defaultWatcher) newTicker() *time.Ticker {
	return time.NewTicker(time.Duration(w.pingInterval+rand.Intn(2000)) * time.Millisecond)
}
