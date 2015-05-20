package server

import (
	"fmt"
)

type Node struct {
	host string
	port int
}

type address string

func (n *Node) Address() address {
	return address(fmt.Sprintf("%s:%d", n.host, n.port))
}
