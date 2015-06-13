package watch

//Takes care to regular pull peers to update its link state
// Adds a ticker on each peer added,
// Each watcher is a goroutine sending pooling messages to a channel
// Master goroutine forwards pool requests to peers and awaits answer

// Basic PING PONG Mechanism

import (
	"fmt"
	"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/node"
	"github.com/marcosQuesada/mesh/peer"
	"log"
	"time"
)

type Watcher interface {
	Watch(peer.NodePeer)
}

type defaultWatcher struct {
	commChan     chan updateMessage
	childs       map[*subject]message.Status //state
	exit         chan bool
	pingInterval int
}

//Used as a message from subjects to watcher
type updateMessage struct {
	id          int
	destination *node.Node
	state       message.Status
}

//a child to take care
type subject struct {
	node    node.Node
	exit    chan bool
	ticker  *time.Ticker
	request int //last request id
}

func New(interval int) *defaultWatcher {
	return &defaultWatcher{pingInterval: interval}
}

func (w *defaultWatcher) Watch(p peer.NodePeer) {
	var id = 0

	go func() {
		for {
			select {
			case <-w.exit:
				return
			case msg := <-p.PingChan():
				//if Ping received Return Pong
				ping := msg.(*message.Ping)
				log.Println("--- Received Ping", ping.Id, "from ", ping.From.String())
				pong := &message.Pong{Id: ping.Id, From: ping.To, To: ping.From}
				id = ping.Id + 1
				p.Send(pong)
			}
		}
	}()

	go func() {
		//@TODO: Randomize this
		ticker := time.NewTicker(time.Duration(w.pingInterval) * time.Second)
		for {
			select {
			case <-w.exit:
				return
			case <-ticker.C:
				//Send Ping to Peer
				ping := &message.Ping{Id: id, From: p.From(), To: p.Node()}
				p.Send(ping)
				node := p.Node()
				log.Println("Sended Ticker PING ", ping.Id, "from ", ping.From.String(), "to ", node.String())
				select {
				case msg := <-p.PongChan():
					pong := msg.(*message.Pong)
					if pong.Id != ping.Id {
						log.Println("Unexpected ID pong", pong.Id, " ping", ping.Id)
					}
					log.Println("Received Pong ", pong.Id, "from ", pong.From.String())

				case <-time.NewTimer(time.Second * 1).C:
					fmt.Println("Waiting Pong: Timeout from ", node.String())
					fmt.Println("-- Declare Dead Pear ", node.String())

					p.Exit()
				}
			}
		}
	}()
}
