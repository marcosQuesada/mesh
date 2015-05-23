package server

import (
	"fmt"
)

type Node struct {
	host string
	port int
}

func (n *Node) String() string {
	return fmt.Sprintf("%s:%d", n.host, n.port)
}
