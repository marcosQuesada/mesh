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
	"github.com/marcosQuesada/mesh/node"
	//"github.com/marcosQuesada/mesh/peer"
	"log"
	"time"
)

type Watcher interface {
	Watch(node.Node)
}

type defaultWatcher struct {
	commChan chan updateMessage
	childs   map[*subject]message.Status //state
	exit     chan bool
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

func New() *defaultWatcher {
	return &defaultWatcher{}
}

func (w *defaultWatcher) OnPeerConnectedEvent(e dispatcher.Event) {
	//	n := e.(*peer.OnPeerConnectedEvent)
	log.Println("Called Watcher OnPeerConnectedEvent", e) //n.Node.String()

}

//func (w *defaultWatcher) Watch(p peer.Peer) {
func (w *defaultWatcher) Watch(n node.Node) {
	/*	s := &subject{
		peer:   p,
		exit:   make(chan bool),
		ticker: time.NewTicker(time.Duration(1) * time.Millisecond),
	}*/

	//w.childs[s] = PeerStatusUnknown
}

func (w *defaultWatcher) Run() {
	for {
		select {
		case <-w.exit:
			return
		case msg := <-w.commChan:
			//check response received and update states
			fmt.Println("msg ", msg)
			//Pong message expected!
			//check that
		}
	}
}

func (t *subject) Watch() {
	for {
		select {
		case <-t.exit:
			return
		case <-t.ticker.C:
			//Send Ping to Peer
			//ping := &Ping{}

			//t.peer.SendPing(ping)
			//case msg := <-t.peer.ReceivePong(): //as chan of messages
			//Pong message expected!
			//check that
		}
	}
}
