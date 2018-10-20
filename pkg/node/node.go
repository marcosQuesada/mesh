package node

import (
	"fmt"
	"strconv"
	"strings"
)

type Node struct {
	Host string
	Port int
}

func (n *Node) String() string {
	return fmt.Sprintf("%s:%d", n.Host, n.Port)
}

func ParseNodeString(nds string) (*Node, error) {
	parts := strings.Split(nds, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("Bad Nodestring!")
	}
	port, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return nil, err
	}

	return &Node{Host: parts[0], Port: int(port)}, nil
}
