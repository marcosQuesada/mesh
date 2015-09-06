package cluster

import (
	"github.com/marcosQuesada/mesh/node"
	"time"
)

type Cluster struct{
	leader node.Node
	leaderLastSeen time.Time
	members map[string] node.Node
}

func (c *Cluster) Start() {

}

func (c *Cluster) Run() {

}

func (c *Cluster) Stop() {

}
