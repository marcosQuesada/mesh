package raft
import "github.com/marcosQuesada/mesh/node"

// Request Types
type VoteRequest struct {
	Candidate node.Node
}

func (v *VoteRequest) action() string {
	return "voteRequest"
}

type PingRequest struct {
	Leader node.Node
}

func (v *PingRequest) action() string {
	return "pingRequest"
}

func NewRaftAction(action string) raftAction {
	switch action {
	case "voteRequest":
		return &VoteRequest{}
	case "pingRequest":
		return &PingRequest{}

	}
	return nil
}