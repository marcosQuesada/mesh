package server

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

type Config struct {
	addr    Node
	cluster map[string]Node
}

func NewConfig(addr, cluster string) *Config {
	addrNode, err := parse(addr)
	if err != nil {
		log.Println("Error parsing Local Address ", addr)

		return nil
	}

	return &Config{
		addr:    addrNode,
		cluster: parseList(cluster),
	}
}

func parseList(clusterList string) map[string]Node {
	parts := strings.Split(clusterList, ",")
	r := make(map[string]Node, len(parts))
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

func parse(node string) (Node, error) {
	nodeParts := clear(node)
	if len(nodeParts) != 0 {
		node = nodeParts[0]
	}
	p := strings.Split(node, ":")
	if len(p) != 2 {
		log.Println("Error building Node on config: ", p)
		return Node{}, fmt.Errorf("Error building Node on config: %s", p)
	}
	port, err := strconv.ParseInt(p[1], 10, 64)
	if err != nil {
		log.Println("Error Parsing Port config: ", p, err)
		return Node{}, err
	}

	return Node{host: p[0], port: int(port)}, nil
}

func clear(node string) []string {
	return strings.Split(node, "\n")
}
