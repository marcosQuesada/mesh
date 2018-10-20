package raft

import (
	"github.com/marcosQuesada/mesh/pkg/message"
	"github.com/marcosQuesada/mesh/pkg/peer"
	"github.com/marcosQuesada/mesh/pkg/router/handler"
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

func (f *FSM) Transactions() map[message.MsgType]bool {
	return map[message.MsgType]bool{
		message.RAFTVOTEREQUEST:      true,
		message.RAFTVOTERESPONSE:     false,
		message.RAFTHEARTBEATREQUEST: false,
	}
}

func (r *Raft) HandleRaftVoteRequest(p peer.PeerNode, msg message.Message) (message.Message, error) {
	cmdMsg := msg.(*message.RaftVoteRequest)
	responseChn := make(chan interface{}, 1)
	r.rcvChan <- handler.Request{ResponseChan: responseChn, Msg: cmdMsg}
	result := <-responseChn

	r.termMutex.Lock()
	term := r.currentTerm
	r.termMutex.Unlock()

	return &message.RaftVoteResponse{
		Id:          cmdMsg.Id,
		From:        r.node,
		Term:        term,
		VoteGranted: result.(bool),
	}, nil
}

//Handled using request listeners on coordinator poolRequest
func (r *Raft) HandleRaftVoteResponse(p peer.PeerNode, msg message.Message) (message.Message, error) {
	return nil, nil
}

func (r *Raft) HandleRaftHeartBeat(p peer.PeerNode, msg message.Message) (message.Message, error) {
	r.rcvChan <- msg.(*message.RaftHeartBeatRequest)

	return nil, nil
}
