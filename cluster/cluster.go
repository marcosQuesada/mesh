package cluster

import (
	"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/node"
	"time"
)

const (
	ClusterStatusStarting  = message.Status("starting")
	ClusterStatusInService = message.Status("in service")
	ClusterStatusDegraded  = message.Status("degraded")
	ClusterStatusExit      = message.Status("exit")
)

type LeadHandler interface {
	Vote(message.Message)
	Candidate(message.Message)
	Affirm(message.Message)
}

type Manager struct {
	leader         node.Node
	leaderLastSeen time.Time
	neighbours     map[string]node.Node
	status         message.Status
}

func NewManager() *Manager {
	return &Manager{
	//leaderLastSeen: time.Now(),
	}
}
func (c *Manager) Start() {

}

func (c *Manager) Run() {

}

func (c *Manager) Stop() {

}

/// Handles Master Election
