package raft
import (
	"github.com/marcosQuesada/mesh/peer"
	"github.com/marcosQuesada/mesh/message"
	"log"
)

func (r *Raft) HandleRaftCommand(p peer.NodePeer, msg message.Message) (message.Message, error) {
	from := p.From()
	cmd := msg.(*message.Command)

	log.Println("HandleCommand, from peer", from.String(), cmd.Command)

	log.Println("HandleCommand Forward request to raft")
	responseChn := make(chan interface{})
	r.rcvChan <- RaftRequest{responseChn, cmd}

	log.Println("HandleCommand Waiting result")

	result := <-responseChn
	log.Println("HandleCommand Forward result", result)

	return message.Response{Id: msg.(*message.Command).Id, From: r.node, Result: result}, nil
}
