package cluster

import (
	"log"

	"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/peer"
	"github.com/marcosQuesada/mesh/router/handler"
)

func (r *Coordinator) Handlers() map[message.MsgType]handler.Handler {
	return map[message.MsgType]handler.Handler{
		message.COMMAND: r.HandleCommand,
	}
}

func (r *Coordinator) HandleCommand(c peer.NodePeer, msg message.Message) (message.Message, error) {
	from := c.From()
	log.Println("HandleCommand, from peer", from.String())

	return nil, nil
}
