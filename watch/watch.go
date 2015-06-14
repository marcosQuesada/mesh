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
	pingInterval int
	mutex        sync.Mutex
}

type subject struct {
	peer   peer.NodePeer
	ticker *time.Ticker
	id     int
}

func New(evCh chan dispatcher.Event, interval int) *defaultWatcher {
	return &defaultWatcher{
		eventChan:    evCh,
		exit:         make(chan bool, 0),
		pingInterval: interval,
	}
}

func (w *defaultWatcher) Watch(p peer.NodePeer) {
	//@TODO: Randomize this
	tickerReset := make(chan bool, 0)
	ticker := newTicker(w.pingInterval)
	s := &subject{p, ticker, 0}
	go w.watchPingChan(s, tickerReset)

	timeout := time.NewTimer(time.Second * 3)
	for {
		select {
		case <-tickerReset:
			log.Println("Reseting ticker")
			s.ticker.Stop()
			s.ticker = newTicker(w.pingInterval)
		case <-s.ticker.C:
			node := p.Node()
			ping := &message.Ping{Id: s.id, From: p.From(), To: node}
			log.Println("Sended Ticker PING ", ping.Id, "from ", ping.From.String(), "to ", node.String())
			p.Send(ping)

			select {
			case msg := <-p.PongChan():
				pong := msg.(*message.Pong)
				if pong.Id != ping.Id {
					w.exit <- true
				}
				log.Println("Received Pong ", pong.Id, "from ", pong.From.String())

				w.mutex.Lock()
				s.id = ping.Id + 1
				w.mutex.Unlock()
				//remove timeout
				timeout.Stop()

			case <-timeout.C:
				log.Println("Timeout waiting pong from: ", node.String(), "Declare Dead")

				w.eventChan <- &peer.OnPeerDisconnectedEvent{
					Node:  p.Node(),
					Event: peer.PeerStatusDisconnected,
					Peer:  p,
				}

				s.peer.Exit()
				w.exit <- true

				return
			}
		case <-w.exit:
			return
		default:
		}
	}
}

func (w *defaultWatcher) Exit() {
	w.exit <- true
}

func (w *defaultWatcher) watchPingChan(s *subject, tickerReset chan bool) {
	for {
		select {
		case <-w.exit:
			return
		case msg := <-s.peer.PingChan():
			//if Ping received Return Pong
			ping := msg.(*message.Ping)
			log.Println("--- Received Ping", ping.Id, "from ", ping.From.String(), "on: ", ping.To.String())
			pong := &message.Pong{Id: ping.Id, From: ping.To, To: ping.From}
			s.peer.Send(pong)

			w.mutex.Lock()
			s.id = ping.Id + 1
			w.mutex.Unlock()

			log.Println("--- Sended Pong", pong.Id, "to ", ping.From.String())
			tickerReset <- true
		}
	}
}
func newTicker(pingInterval int) *time.Ticker {
	return time.NewTicker(time.Duration(pingInterval) * time.Second)
}
