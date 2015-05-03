package cluster

import (
	"fmt"
	"time"
)

type Pool struct {
	peers []Client
}

func (p *Pool) Add(peer Client) {
	p.peers = append(p.peers, peer)
}

func (p *Pool) Monitor() {
	for {
		time.Sleep(1 * time.Second)
		//sends heartbeat to peer 1
	}
}
func init() {
	fmt.Println("Coordinator init")
}
