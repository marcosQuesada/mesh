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
	go func() {
		//@TODO: Randomize this
		ticker := time.NewTicker(time.Duration(5) * time.Second)
		var id = 0
		for {
			select {
			case <-w.exit:
				return

			case <-ticker.C:
				//Send Ping to Peer
				ping := &message.Ping{Id: id, From: p.From(), To: p.Node()}
				p.Send(ping)
				fmt.Println("Sended PING msg ", ping)
				//Pong response must be handled using timeouts to detect link failures

			case msg := <-p.PingChan():
				fmt.Println("Received From PingChan ", msg)
				//if Ping received Return Pong
				switch msg.(type) {
				case *message.Ping:
					log.Println("Received Ping", msg)
					ping := msg.(*message.Ping)
					id = ping.Id + 1
					pong := &message.Pong{Id: id, From: ping.To, To: ping.From}
					p.Send(pong)
					fmt.Println("Sended PONG msg ", pong)

				case *message.Pong:
					pong := msg.(*message.Pong)

					log.Println("Received Pong ", pong.Id)

				}
			}
		}
	}()
}
