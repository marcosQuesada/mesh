package raft

import (
	"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/peer"
	"github.com/marcosQuesada/mesh/router/handler"
	//"log"
)

func (f *FSM) Handlers() map[message.MsgType]handler.Handler {
	r, _ := f.State.(*Raft)

	return map[message.MsgType]handler.Handler{
		message.RAFTVOTEREQUEST:      r.HandleRaftVoteRequest,
		message.RAFTVOTERESPONSE:     r.HandleRaftVoteResponse,
		message.RAFTHEARTBEATREQUEST: r.HandleRaftHeartBeat,
	}
}
func (f *FSM) Notifiers() map[message.MsgType]bool {
	return map[message.MsgType]bool{
		message.RAFTVOTEREQUEST:      false,
		message.RAFTVOTERESPONSE:     true,
		message.RAFTHEARTBEATREQUEST: false,
	}
}

func (r *Raft) HandleRaftVoteRequest(p peer.NodePeer, msg message.Message) (message.Message, error) {
	cmdMsg := msg.(*message.RaftVoteRequest)

	//log.Println("HandleRaftVoteRequest, from peer", cmdMsg.From.String(), "candidate: ", cmdMsg.Candidate.String())

	responseChn := make(chan interface{}, 1)
	r.rcvChan <-handler.Request{ResponseChan:responseChn, Msg:cmdMsg}
	result := <-responseChn

	return &message.RaftVoteResponse{Id: cmdMsg.Id, From: r.node, Vote: result.(string)}, nil
}

func (r *Raft) HandleRaftVoteResponse(p peer.NodePeer, msg message.Message) (message.Message, error) {
	//cmdMsg := msg.(*message.RaftVoteResponse)

	//log.Println("HandleRaftVoteResponse, from peer", cmdMsg.From.String(), "vote: ", cmdMsg.Vote)

	return nil, nil
}

func (r *Raft) HandleRaftHeartBeat(p peer.NodePeer, msg message.Message) (message.Message, error) {
	cmdMsg := msg.(*message.RaftHeartBeatRequest)

	//log.Println("HandleRaftHeartBeat, from peer", cmdMsg.From.String(), "leader: ", cmdMsg.Leader.String())

	r.rcvChan <- cmdMsg

	return nil, nil
}
