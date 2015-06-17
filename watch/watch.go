package watch

//Takes care to regular pull peers to update its link state
// Adds a ticker on each peer added,
// Each watcher is a goroutine sending pooling messages to a channel
// Master goroutine forwards pool requests to peers and awaits answer

// Basic PING PONG Mechanism

import (
	"github.com/marcosQuesada/mesh/dispatcher"
	"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/peer"
	"log"
	"sync"
	"time"
)

type Watcher interface {
	Watch(peer.NodePeer)
}

type defaultWatcher struct {
	eventChan    chan dispatcher.Event
	exit         chan bool
	done         chan bool
	pingInterval int
	mutex        sync.RWMutex
	index        map[string]bool
}

type subject struct {
	peer   peer.NodePeer
	ticker *time.Ticker
	id     int
}

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}
func New(evCh chan dispatcher.Event, interval int) *defaultWatcher {
	return &defaultWatcher{
		eventChan:    evCh,
		exit:         make(chan bool, 0),
		done:         make(chan bool, 1),
		index:        make(map[string]bool, 0),
		pingInterval: interval,
	}
}

func (w *defaultWatcher) Watch(p peer.NodePeer) {
	defer log.Println("Watch Exits")
	node := p.Node()
	w.mutex.Lock()
	w.index[node.String()] = true
	w.mutex.Unlock()

	//@TODO: Randomize this
	tickerReset := make(chan bool, 1)
	ticker := newTicker(w.pingInterval)
	s := &subject{p, ticker, 0}
	go w.watchPingChan(s, tickerReset)

	var timeout *time.Timer
	for {
		select {
		case <-tickerReset:
			log.Println("Reseting ticker")
			/*			s.ticker.Stop()
						s.ticker = newTicker(w.pingInterval)*/
		case <-s.ticker.C:
			log.Println("Fire Ticker")
			//node := p.Node()
			ping := &message.Ping{Id: s.id, From: p.From(), To: node}
			log.Println("Sended Ticker PING ", s.id, "from ", ping.From.String(), "to ", node.String())
			p.Send(ping)
			timeout = time.NewTimer(time.Second * 3)

			select {
			case msg, open := <-p.PongChan():
				if !open {
					log.Print("PongChan closed, exit")
					return
				}
				//remove timeout
				timeout.Stop()
				pong := msg.(*message.Pong)
				if pong.Id != ping.Id {
					w.exit <- true
				}
				log.Println("Received Pong ", pong.Id, "from ", pong.From.String())

				w.mutex.Lock()
				s.id = ping.Id + 1
				w.mutex.Unlock()

			case <-timeout.C:
				log.Println("Timeout waiting pong from: ", node.String(), "id:", s.id, "Declare Dead")

				w.eventChan <- &peer.OnPeerDisconnectedEvent{
					Node:  p.Node(),
					Event: peer.PeerStatusDisconnected,
					Peer:  p,
				}

				s.peer.Exit()
				w.exit <- true

				return
			case <-w.exit:
				timeout.Stop()
				w.done <- true
				return

			}
		case <-w.exit:
			//			timeout.Stop()
			w.done <- true
			return
			//default:
		}
	}

}

func (w *defaultWatcher) Exit() {
	w.exit <- true
	close(w.exit)
	for range w.index {
		<-w.done
		log.Println("Stoping Watchers")
	}
	close(w.done)
	log.Println("Exiting Done")
}

func (w *defaultWatcher) watchPingChan(s *subject, tickerReset chan bool) {
	for {
		select {
		case <-w.exit:
			return
		case msg, open := <-s.peer.PingChan():
			if !open {
				log.Println("Ping Channel closed, exit")
				return
			}
			tickerReset <- true
			//if Ping received Return Pong
			ping := msg.(*message.Ping)
			log.Println("--- Received Ping", ping.Id, "from ", ping.From.String(), "on: ", ping.To.String())
			pong := &message.Pong{Id: ping.Id, From: ping.To, To: ping.From}
			s.peer.Send(pong)

			w.mutex.Lock()
			s.id = ping.Id + 1
			w.mutex.Unlock()

			log.Println("--- Sended Pong", pong.Id, "to ", ping.From.String())
			/*					log.Println("Reseting ticker")
			s.ticker.Stop()
			s.ticker = newTicker(w.pingInterval)*/
		}
	}
}
func newTicker(pingInterval int) *time.Ticker {
	return time.NewTicker(time.Duration(pingInterval) * time.Second)
}
