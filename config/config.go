package config

import (
	"fmt"
	n "github.com/marcosQuesada/mesh/node"
	"log"
	"strconv"
	"strings"
)

type Config struct {
	Addr    n.Node
	Cluster map[string]n.Node
}

func NewConfig(addr, cluster string) *Config {
	addrNode, err := parse(addr)
	if err != nil {
		log.Println("Error parsing Local Address ", addr)

		return nil
	}

	return &Config{
		Addr:    addrNode,
		Cluster: parseList(cluster),
	}
}

func parseList(clusterList string) map[string]n.Node {
	parts := strings.Split(clusterList, ",")
	r := make(map[string]n.Node, len(parts))
	if len(parts) > 0 && len(parts[0]) > 0 {
		log.Println("Parts are: ", parts)
		for _, nodePart := range parts {
			node, err := parse(nodePart)
			if err != nil {
				continue
			}
			r[node.String()] = node
		}
		return r
	}

	return nil
}

func parse(node string) (n.Node, error) {
	nodeParts := clear(node)
	if len(nodeParts) != 0 {
		node = nodeParts[0]
	}
	p := strings.Split(node, ":")
	if len(p) != 2 {
		log.Println("Error building Node on config: ", p)
		return n.Node{}, fmt.Errorf("Error building Node on config: %s", p)
	}
	port, err := strconv.ParseInt(p[1], 10, 64)
	if err != nil {
		log.Println("Error Parsing Port config: ", p, err)
		return n.Node{}, err
	}

	return n.Node{Host: p[0], Port: int(port)}, nil
}

func clear(node string) []string {
	return strings.Split(node, "\n")
}
