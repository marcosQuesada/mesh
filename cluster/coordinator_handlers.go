package cluster

import (
	"log"

	"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/peer"
	"github.com/marcosQuesada/mesh/router/handler"
)

func (r *Coordinator) Handlers() map[message.MsgType]handler.Handler {
	return map[message.MsgType]handler.Handler{
		message.COMMAND:  r.HandleCommand,
		message.RESPONSE: r.HandleResponse,
	}
}

func (r *Coordinator) Notifiers() map[message.MsgType]bool {
	return map[message.MsgType]bool{
		message.COMMAND:  false,
		message.RESPONSE: true,
	}
}

func (r *Coordinator) HandleCommand(c peer.NodePeer, msg message.Message) (message.Message, error) {
	from := c.From()
	log.Println("HandleCommand, from peer", from.String())

	return nil, nil
}

func (r *Coordinator) HandleResponse(c peer.NodePeer, msg message.Message) (message.Message, error) {
	from := c.From()
	log.Println("HandleResponse, from peer", from.String())

	return nil, nil
}
