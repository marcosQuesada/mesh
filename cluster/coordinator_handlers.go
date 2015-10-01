package cluster

import (
	"github.com/marcosQuesada/mesh/cluster/raft"
	"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/node"
	"github.com/marcosQuesada/mesh/peer"
	"github.com/marcosQuesada/mesh/router/handler"
	"github.com/mitchellh/mapstructure"
	"log"
	"reflect"
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

func (c *Coordinator) HandleCommand(p peer.NodePeer, msg message.Message) (message.Message, error) {
	cmdMsg := msg.(*message.Command)
	rawCommand := cmdMsg.Command.(map[string]interface{})

	//@TODO : ADOPT INTERFACES!
	cmd := &raft.VoteRequest{Candidate: node.Node{}}
	err := mapstructure.Decode(rawCommand, cmd)
	if err != nil {
		log.Println("Error decoding command ")
		return nil, err
	}

	var res interface{} = cmd

	if cmd.Candidate == (node.Node{}) {
		ping := &raft.PingRequest{Leader: node.Node{}}
		err = mapstructure.Decode(rawCommand, ping)
		if err != nil {
			log.Println("Error decoding command ")
			return nil, err
		}
		res = ping

	}

	log.Println("HandleCommand, from peer", cmdMsg.From.String(), reflect.TypeOf(res).String())
	rspChan := c.manager.Response()
	responseChn := make(chan interface{}, 1)
	rspChan <- raft.RaftRequest{responseChn, res}
	result := <-responseChn

	return message.Response{Id: cmdMsg.Id, From: c.from, Result: result}, nil
}

//@TODO: Rethink!!! request Listeners are handling responses by itselfs
func (c *Coordinator) HandleResponse(p peer.NodePeer, msg message.Message) (message.Message, error) {
	m := msg.(*message.Response)
	log.Println("HandleResponse, from peer", m.From.String(), msg.ID())

	return nil, nil
}
