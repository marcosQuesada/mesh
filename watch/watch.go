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
}

type defaultWatcher struct {
	eventChan    chan dispatcher.Event
	exit         chan bool
	pingInterval int
	mutex        sync.RWMutex
	index        map[string]*subject
	wg           sync.WaitGroup
}

type subject struct {
	peer        peer.NodePeer
	ticker      *time.Ticker
	id          int
	Done        chan bool
	tickerReset chan bool
}

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
}
func New(evCh chan dispatcher.Event, interval int) *defaultWatcher {
	return &defaultWatcher{
		eventChan:    evCh,
		exit:         make(chan bool, 0),
		index:        make(map[string]*subject, 0),
		pingInterval: interval,
	}
}

func (w *defaultWatcher) Watch(p peer.NodePeer) {
	defer log.Println("Watcher", p.Node(), "Exits")

	//add watcher to waitGroup
	w.wg.Add(1)
	defer w.wg.Done()

	//@TODO: Randomize this
	tickerReset := make(chan bool, 10)
	ticker := newTicker(w.pingInterval)
	subjectDone := make(chan bool, 10)
	s := &subject{p, ticker, 0, subjectDone, tickerReset}

	node := p.Node()
	w.mutex.Lock()
	w.index[node.String()] = s
	w.mutex.Unlock()

	go w.watchPingChan(s)

	var timeout *time.Timer
	for {
		timeout = time.NewTimer(time.Second * 3)
		select {
		case <-s.tickerReset:
			log.Println("Reseting ticker", s.id, "to", p.Node())

			//@TODO: TEST!

			s.ticker.Stop()
			timeout.Stop()
			s.ticker = newTicker(w.pingInterval)
		case <-s.ticker.C:
			go func() {
				ping := &message.Ping{Id: s.id, From: p.From(), To: node}
				log.Println("Ticker PING", s.id, "to", node.String())
				p.Send(ping)

				//timeout = time.NewTimer(time.Second * 3)
				select {
				case msg, open := <-p.PongChan():
					if !open {
						log.Print("PongChan closed, exit")

						w.eventChan <- &peer.OnPeerDisconnectedEvent{
							Node:  p.Node(),
							Event: peer.PeerStatusDisconnected,
							Peer:  p,
						}

						s.Done <- true
						return
					}
					//remove timeout
					timeout.Stop()
					pong := msg.(*message.Pong)
					if pong.Id != ping.Id {
						log.Println("Mistmatched IDs!!!!!", pong.Id, ping.Id)
					}
					log.Println("Received Pong", pong.Id, "from", pong.From.String())

					w.mutex.Lock()
					s.id = ping.Id + 1
					w.mutex.Unlock()

				case <-timeout.C:
					log.Println("Timeout id:", s.id, "IsDead", node.String())

					w.eventChan <- &peer.OnPeerDisconnectedEvent{
						Node:  p.Node(),
						Event: peer.PeerStatusDisconnected,
						Peer:  p,
					}

					s.peer.Exit()

					return

				//Watcher Done inside Pong Expectation
				case <-s.Done:
					timeout.Stop()

					return
				}
			}()

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

func (w *defaultWatcher) watchPingChan(s *subject) {
	defer log.Println("Exiting WatchPingChan ", s.peer.From())
	for {
		select {
		case <-w.exit:
			return
		case msg, open := <-s.peer.PingChan():
			if !open {
				log.Println("Ping Channel closed, exit")
				return
			}

			//Reset ticker
			s.tickerReset <- true

			//if Ping received Return Pong
			ping := msg.(*message.Ping)
			log.Println("Received Ping", ping.Id, "on: ", ping.From.String())
			pong := &message.Pong{Id: ping.Id, From: ping.To, To: ping.From}
			s.peer.Send(pong)

			/*			w.mutex.Lock()
						s.id = ping.Id + 1
						w.mutex.Unlock()*/

			log.Println("--Sended Pong", pong.Id, "to ", ping.From.String())
		}
	}
}
func newTicker(pingInterval int) *time.Ticker {
	return time.NewTicker(time.Duration(pingInterval) * time.Second)
}
