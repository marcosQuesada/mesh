package watch

// Takes care to regular pull peers to update its link state
// Adds a ticker on each peer added,
// Each watcher is a goroutine sending pooling messages to a channel
// Master goroutine forwards pool requests to peers and awaits answer

// Basic PING PONG Mechanism

import (
	"log"
	"math/rand"
	"sync"
	"time"

	"errors"
	"github.com/marcosQuesada/mesh/pkg/dispatcher"
	"github.com/marcosQuesada/mesh/pkg/message"
	"github.com/marcosQuesada/mesh/pkg/peer"
	"github.com/marcosQuesada/mesh/pkg/router/request"
)

type Watcher interface {
	Watch(peer.PeerNode)
}

type subject struct {
	peer     peer.PeerNode
	ticker   *time.Ticker
	id       message.ID
	Done     chan bool
	mutex    sync.RWMutex
	lastSeen time.Time
}

func (s *subject) incId() { //@TODO: TO atomic pointers
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.id = message.NewId()
}

func (s *subject) getId() message.ID {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.id
}

type defaultWatcher struct {
	eventChan       chan dispatcher.Event
	exit            chan bool
	pingInterval    int
	mutex           sync.RWMutex
	index           map[string]*subject
	requestListener *request.RequestListener
	wg              sync.WaitGroup
}

func New(rqLst *request.RequestListener, evCh chan dispatcher.Event, interval int) *defaultWatcher {
	rand.Seed(time.Now().UTC().UnixNano())

	return &defaultWatcher{
		eventChan:       evCh,
		exit:            make(chan bool, 0),
		index:           make(map[string]*subject, 0),
		pingInterval:    interval * 1000,
		requestListener: rqLst,
	}
}

func (w *defaultWatcher) Watch(p peer.PeerNode) {
	//add watcher to waitGroup
	w.wg.Add(1)
	defer w.wg.Done()

	subjectDone := make(chan bool, 10)
	s := &subject{
		peer:     p,
		ticker:   w.newTicker(),
		Done:     subjectDone,
		id:       message.NewId(),
		lastSeen: time.Now(),
	}
	defer close(s.Done)

	node := p.Node()
	w.mutex.Lock()
	w.index[node.String()] = s
	w.mutex.Unlock()

	for {
		select {
		case _, open := <-s.peer.ResetWatcherChan():
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

			msg, err := w.requestListener.RegisterAndWait(requestId)
			s.incId()
			s.ticker = w.newTicker()
			// @TODO: Handle response error
			if err != nil {
				log.Println("XXXXX Error Waiting Response ", err)
				continue
			}

			if msg.MessageType() != message.PONG {
				log.Println("Error Unexpected Received type, expected PONG ", msg.MessageType(), "RequestListener ", requestId)

				//@TODO: used to check development stability
				panic(errors.New("PONG CRASH"))
			}
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
