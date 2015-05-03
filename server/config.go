package server

import (
	"fmt"
	"strconv"
	"strings"
)

type Config struct {
	addr          string
	raft_addr     string
	raft_data_dir string
	raft_cluster  []*Node
}

func NewConfig(addr, raftAddr, raftCluster, raftDataDir string) *Config {
	return &Config{
		addr:          addr,
		raft_addr:     raftAddr,
		raft_data_dir: raftDataDir,
		raft_cluster:  parse(raftCluster),
	}
}

type Node struct {
	host string
	port int
}

func (n *Node) Address() string {
	return fmt.Sprintf("%s:%d", n.host, n.port)
}

func parse(raftCluster string) []*Node {
	parts := strings.Split(raftCluster, ",")
	if len(parts) > 0 && len(parts[0]) > 0 {
		fmt.Println("Parts are: ", parts)
		var nodes []*Node
		for _, nodePart := range parts {
			p := strings.Split(nodePart, ":")
			if len(p) != 2 {
				fmt.Println("Error building Node on config: ", p)
				continue
			}
			port, err := strconv.ParseInt(p[1], 10, 64)
			if err != nil {
				fmt.Println("Error Parsing Port config: ", p)
				continue
			}
			nodes = append(nodes, &Node{host: p[0], port: int(port)})
		}
		return nodes
	}

	return nil
}
