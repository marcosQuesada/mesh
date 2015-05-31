package server

import (
	"fmt"
)

type Node struct {
	Host string
	Port int
}

func (n *Node) String() string {
	return fmt.Sprintf("%s:%d", n.Host, n.Port)
}
