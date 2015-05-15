package server

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

type Config struct {
	addr          *Node
	raft_addr     *Node
	raft_data_dir string
	raft_cluster  []*Node
}
type address string

func NewConfig(addr, raftAddr, raftCluster, raftDataDir string) *Config {
	addrNode, err := parse(addr)
	if err != nil {
		log.Println("Error parsing Local Address ", addr)

		return nil
	}

	raftAddrNode, err := parse(raftAddr)
	if err != nil {
		log.Println("Error parsing Raft Address ", raftAddrNode)

		return nil
	}

	return &Config{
		addr:          addrNode,
		raft_addr:     raftAddrNode,
		raft_data_dir: raftDataDir,
		raft_cluster:  parseList(raftCluster),
	}
}

type Node struct {
	host string
	port int
}

func (n *Node) Address() address {
	return address(fmt.Sprintf("%s:%d", n.host, n.port))
}

func parseList(raftCluster string) []*Node {
	parts := strings.Split(raftCluster, ",")
	if len(parts) > 0 && len(parts[0]) > 0 {
		log.Println("Parts are: ", parts)
		var nodes []*Node
		for _, nodePart := range parts {
			node, err := parse(nodePart)
			if err != nil {
				continue
			}
			nodes = append(nodes, node)
		}
		return nodes
	}

	return nil
}

func parse(node string) (*Node, error) {
	nodeParts := clear(node)
	if len(nodeParts) != 0 {
		node = nodeParts[0]
	}
	p := strings.Split(node, ":")
	if len(p) != 2 {
		log.Println("Error building Node on config: ", p)
		return nil, fmt.Errorf("Error building Node on config: %s", p)
	}
	port, err := strconv.ParseInt(p[1], 10, 64)
	if err != nil {
		log.Println("Error Parsing Port config: ", p, err)
		return nil, err
	}

	return &Node{host: p[0], port: int(port)}, nil
}

func clear(node string) []string {
	return strings.Split(node, "\n")
}
