package cluster

import (
	"log"

	"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/peer"
)

func (r *Coordinator) HandleCommand(c peer.NodePeer, msg message.Message) (message.Message, error) {
	log.Println("HandleCommand, from peer", c.From().String())

	return nil, nil
}
