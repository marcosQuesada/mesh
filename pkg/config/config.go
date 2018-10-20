package config

import (
	"fmt"
	n "github.com/marcosQuesada/mesh/pkg/node"
	yml "github.com/olebedev/config"
	"log"
	"strconv"
	"strings"
)

type Config struct {
	Addr    n.Node
	Cluster map[string]n.Node
	CliPort int
}

func NewConfig(addr, cluster string, cliPort int) *Config {
	addrNode, err := parse(addr)
	if err != nil {
		log.Println("Error parsing Local Address ", addr)

		return nil
	}

	return &Config{
		Addr:    addrNode,
		Cluster: parseList(cluster),
		CliPort: cliPort,
	}
}

func ParseYML(cfgFile string) *Config {
	cfg, err := yml.ParseYamlFile(cfgFile)
	if err != nil {
		log.Println("Unexpected Error Parsing Yml", err)
		return nil
	}

	addrNode, err := cfg.String("cluster.node")
	if err != nil {
		log.Println("Unexpected Error Parsing cluster.node", err)
		return nil
	}
	localNode, err := parse(addrNode)
	if err != nil {
		log.Println("Unexpected Error Parsing LocalNode", err)
		return nil
	}
	memberList, err := cfg.List("cluster.members")
	if err != nil {
		log.Println("Unexpected Error Parsing cluster.members", err)
		return nil
	}

	members := make(map[string]n.Node, len(memberList))
	for _, nodePart := range memberList {
		node, err := parse(nodePart.(string))
		if err != nil {
			continue
		}
		members[node.String()] = node
	}

	cliPort, err := cfg.Int("cli.port")
	if err != nil {
		log.Println("Unexpected Error Parsing cli.port", err)
		return nil
	}

	return &Config{
		Addr:    localNode,
		Cluster: members,
		CliPort: cliPort,
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
