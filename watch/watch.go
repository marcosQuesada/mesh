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
	log.Println("XXX Watch from ", p.From(), "to", p.Node())
	var id = 0
	go func(c peer.NodePeer) {
		log.Println("XXX---XXX Watch from ", c.From(), "to", c.Node())
		//@TODO: Randomize this
		ticker := newTicker(w.pingInterval)
		for {
			select {
			case <-w.exit:
				return
			case msg := <-p.PingChan():
				//if Ping received Return Pong
				ping := msg.(*message.Ping)
				log.Println("--- Received Ping", ping.Id, "from ", ping.From.String())
				pong := &message.Pong{Id: ping.Id, From: ping.To, To: ping.From}

				w.mutex.Lock()
				id = ping.Id + 1
				w.mutex.Unlock()

				p.Send(pong)
				log.Println("--- Sended Pong", pong.Id, "to ", ping.From.String())
				ticker.Stop()
				ticker = newTicker(w.pingInterval)
				log.Println("--- Restarting ticker")

			case <-ticker.C:
				//Send Ping to Peer
				ping := &message.Ping{Id: id, From: c.From(), To: c.Node()}
				c.Send(ping)
				node := c.Node()
				log.Println("Sended Ticker PING ", ping.Id, "from ", ping.From.String(), "to ", node.String())
				select {
				case msg := <-c.PongChan():
					pong := msg.(*message.Pong)
					if pong.Id != ping.Id {
						log.Println("XXX Unexpected ID pong", pong.Id, " ping", ping.Id)
						w.exit <- true
					}
					log.Println("Received Pong ", pong.Id, "from ", pong.From.String())

					w.mutex.Lock()
					id = ping.Id + 1
					w.mutex.Unlock()

				case <-time.NewTimer(time.Second * 5).C:
					fmt.Println("Waiting Pong: Timeout from ", node.String())
					fmt.Println("-- Declare Dead Pear ", node.String())

					w.eventChan <- &peer.OnPeerDisconnectedEvent{
						Node:  c.Node(),
						Event: peer.PeerStatusDisconnected,
						Peer:  p,
					}

					p.Exit()
					w.exit <- true
				}
			}
		}
	}(p)
}

func (w *defaultWatcher) Exit() {
	w.exit <- true
}

func newTicker(pingInterval int) *time.Ticker {
	return time.NewTicker(time.Duration(pingInterval) * time.Second)
}
