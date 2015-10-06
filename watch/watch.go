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
	"errors"
)

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
	lastSeen time.Time
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

func New(rqLst *RequestListener, evCh chan dispatcher.Event, interval int) *defaultWatcher {
	return &defaultWatcher{
		eventChan:       evCh,
		exit:            make(chan bool, 0),
		index:           make(map[string]*subject, 0),
		pingInterval:    interval * 1000,
		requestListener: rqLst,
	}
}

func (w *defaultWatcher) Watch(p peer.NodePeer) {
	//add watcher to waitGroup
	w.wg.Add(1)
	defer w.wg.Done()

	subjectDone := make(chan bool, 10)
	s := &subject{
		peer:   p,
		ticker: w.newTicker(),
		Done:   subjectDone,
		id: message.NewId(),
		lastSeen: time.Now(),
	}
	defer close(s.Done)

	node := p.Node()
	w.mutex.Lock()
	w.index[node.String()] = s
	w.mutex.Unlock()

	for {
		select {
		case _, open := <- s.peer.ResetWatcherChan():
			if !open {
				log.Println("Closed Reset Watcher Chan, exiting watcher")
				return
			}

			s.ticker.Stop()
			s.ticker = w.newTicker()
			continue
		case <-s.ticker.C:
			requestId := s.getId()
			p.Commit(&message.Ping{Id: requestId, From: p.From(), To: node})
			s.ticker.Stop()

			//wait ping response
			msg := w.requestListener.Transaction(requestId)
			if msg.MessageType() != message.PONG {
				log.Println("Error Unexpected Received type, expected PONG ", msg.MessageType(), "RequestListener ", requestId)
				//@TODO: used to check development stability
				panic(errors.New("PONG CRASH"))
			}

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
	return time.NewTicker(time.Duration(w.pingInterval+rand.Intn(10000)) * time.Millisecond)
}
