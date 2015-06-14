package watch

//Takes care to regular pull peers to update its link state
// Adds a ticker on each peer added,
// Each watcher is a goroutine sending pooling messages to a channel
// Master goroutine forwards pool requests to peers and awaits answer

// Basic PING PONG Mechanism

import (
	"fmt"
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

func New(evCh chan dispatcher.Event, interval int) *defaultWatcher {
	return &defaultWatcher{
		eventChan:    evCh,
		exit:         make(chan bool, 0),
		pingInterval: interval,
	}
}

func (w *defaultWatcher) Watch(p peer.NodePeer) {
	defer fmt.Println("Exiting Watch ", p.Node())
	log.Println("Begin")
	var id = 0
	//log.Println("INIT from ", p.From(), "To:", p.Node(), p.From())
	//@TODO: Randomize this
	ticker := newTicker(w.pingInterval)
	timeout := time.NewTimer(time.Second * 5)
	go w.watchPingChan(p, ticker, &id)
	for {
		select {
		case <-ticker.C:
			ping := &message.Ping{Id: id, From: p.From(), To: p.Node()}
			p.Send(ping)

			node := p.Node()
			log.Println("Sended Ticker PING ", ping.Id, "from ", ping.From.String(), "to ", node.String())
			select {
			case msg := <-p.PongChan():
				pong := msg.(*message.Pong)
				if pong.Id != ping.Id {
					log.Println("XXX Unexpected ID pong", pong.Id, " ping", ping.Id)
					w.exit <- true
				}
				log.Println("Received Pong ", pong.Id, "from ", pong.From.String())

				w.mutex.Lock()
				id = ping.Id + 1
				w.mutex.Unlock()
				log.Println("ID is ", id, "stop timeout response")
				timeout.Stop()

			case <-timeout.C:
				fmt.Println("Timeout waiting pong from: ", node.String())
				fmt.Println("-- Declare Dead Pear ", node.String())

				w.eventChan <- &peer.OnPeerDisconnectedEvent{
					Node:  p.Node(),
					Event: peer.PeerStatusDisconnected,
					Peer:  p,
				}

				p.Exit()
				w.exit <- true

				return
			}
		case <-w.exit:
			fmt.Println("Exiting ticker watch ")
			return
		}
	}
}

func (w *defaultWatcher) Exit() {
	w.exit <- true
}

func (w *defaultWatcher) watchPingChan(p peer.NodePeer, ticker *time.Ticker, id *int) {
	defer fmt.Println("Exiting watchPingChan ", p.Node())
	for {
		select {
		case <-w.exit:
			return
		case msg := <-p.PingChan():
			//if Ping received Return Pong
			ping := msg.(*message.Ping)
			log.Println("--- Received Ping", ping.Id, "from ", ping.From.String(), "on: ", ping.To.String())
			pong := &message.Pong{Id: ping.Id, From: ping.To, To: ping.From}
			p.Send(pong)

			w.mutex.Lock()
			*id = ping.Id + 1
			w.mutex.Unlock()

			log.Println("--- Sended Pong", pong.Id, "to ", ping.From.String())

			//ticker.Stop()
			//ticker = newTicker(w.pingInterval)
			//log.Println("--- Restarting ticker ???")
		}
	}
}
func newTicker(pingInterval int) *time.Ticker {
	return time.NewTicker(time.Duration(pingInterval) * time.Second)
}
