package watch

//Watcher handles life peer states
// Start peers on boot
// takes note of last seen
// on timeout uses Ping Pong mechanism
// handles failure detection and Peer reconnection

import (
	"github.com/marcosQuesada/mesh/peer"
)

type Watcher interface {
	Watch(peer.NodePeer)
}

//TODO: Refactor InitDialClient Stuff

