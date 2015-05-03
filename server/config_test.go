package server

import (
	"testing"
)

func TestNodeAddress(t *testing.T) {
	node := &Node{host: "localhost", port: 1234}
	if node.Address() != "localhost:1234" {
		t.Error("Config Error Generating string address")
	}
}

func TestParseOnSuccess(t *testing.T) {
	nodeList := "127.0.0.1:12000,127.0.0.1:12001,127.0.0.1:12002"
	nodes := parseList(nodeList)

	if len(nodes) != 3 {
		t.Error("Bad Result parsing Node List, expected 3, are:", nodes)
	}

	if nodes[0].host != "127.0.0.1" {
		t.Error("Bad Result parsing , Unexpected host", nodes[0])
	}

	if nodes[0].port != 12000 {
		t.Error("Bad Result parsing , Unexpected port", nodes[0])
	}

	if nodes[1].port != 12001 {
		t.Error("Bad Result parsing , Unexpected port", nodes[1])
	}

	if nodes[2].port != 12002 {
		t.Error("Bad Result parsing , Unexpected port", nodes[2])
	}
}

func TestParseOnErrorsList(t *testing.T) {
	nodeList := "127.0.0.1:12001x127.0.0.1:12002"
	nodes := parseList(nodeList)

	if len(nodes) != 0 {
		t.Error("Bad Result parsing Node List, expected 0, are:", nodes)
	}
}

func TestParseOnErrors(t *testing.T) {
	nodeList := "127.0.0.1,127.0.0.1:12001,127.0.0.1:12002"
	nodes := parseList(nodeList)

	if len(nodes) != 2 {
		t.Error("Bad Result parsing Node List, expected 2, are:", nodes)
	}
}
