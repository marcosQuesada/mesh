package watch

//Watcher handles life peer states
// Start peers on boot
// takes note of last seen
// on timeout uses Ping Pong mechanism
// handles failure detection and Peer reconnection

import (
	//"github.com/marcosQuesada/mesh/dispatcher"
	"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/peer"
	//"github.com/marcosQuesada/mesh/router"
	//"log"
)

type Watcher interface {
	Watch(peer.NodePeer)
	HandlePing(peer.NodePeer, message.Message) (message.Message, error)
	HandlePong(peer.NodePeer, message.Message) (message.Message, error)
}

//go r.watcher.Watch(c)
/*

func (r *defaultRouter) InitDialClient(destination node.Node) {
	//Blocking call, wait until connection success
	p := peer.NewDialer(r.from, destination)
	go p.Run()
	log.Println("Connected Dial Client from Node ", r.from.String(), "destination: ", destination.String(), p.Id())

	//Say Hello and wait response
	id, err := p.SayHello()
	if err != nil {
		log.Println("Error getting Hello Id ", err)
	}

	requestId := r.requestListener.Id(p.Node(), id)
	r.requestListener.Register(requestId)

	r.Accept(p)
}
*/
