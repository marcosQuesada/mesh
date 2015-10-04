package cluster

import (
	"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/peer"
	"github.com/marcosQuesada/mesh/router/handler"
	"log"
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

func (r *Coordinator) Transactions() map[message.MsgType]bool {
	return map[message.MsgType]bool{
		message.COMMAND:  true,
		message.RESPONSE: false,
	}
}

func (c *Coordinator) HandleCommand(p peer.NodePeer, msg message.Message) (message.Message, error) {
	cmdMsg := msg.(*message.Command)
	log.Println("HandleCommand, from peer", cmdMsg.From.String(), msg.ID())

	return message.Response{Id: cmdMsg.Id, From: c.from}, nil
}

func (c *Coordinator) HandleResponse(p peer.NodePeer, msg message.Message) (message.Message, error) {
	m := msg.(*message.Response)
	log.Println("HandleResponse, from peer", m.From.String(), msg.ID())

	return nil, nil
}
