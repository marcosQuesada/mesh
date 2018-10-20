package config

import (
	n "github.com/marcosQuesada/mesh/pkg/node"
	yml "github.com/olebedev/config"
	"testing"
)

func TestNodeAddress(t *testing.T) {
	node := &n.Node{Host: "localhost", Port: 1234}
	if node.String() != "localhost:1234" {
		t.Error("Config Error Generating string address")
	}
}

func TestParseOnSuccess(t *testing.T) {
	nodeList := "127.0.0.1:12000,127.0.0.1:12001,127.0.0.1:12002"
	nodes := parseList(nodeList)

	if len(nodes) != 3 {
		t.Error("Bad Result parsing Node List, expected 3, are:", nodes)
	}

	/*	if nodes["127.0.0.1:12000"] == nil {
		t.Error("Bad Result parsing , Unexpected host", nodes["127.0.0.1:12000"])
	}*/

	if nodes["127.0.0.1:12000"].Host != "127.0.0.1" {
		t.Error("Bad Result parsing , Unexpected host", nodes["127.0.0.1:12000"])
	}

	if nodes["127.0.0.1:12000"].Port != 12000 {
		t.Error("Bad Result parsing , Unexpected port", nodes["127.0.0.1:12000"])
	}

	if nodes["127.0.0.1:12001"].Port != 12001 {
		t.Error("Bad Result parsing , Unexpected port", nodes["127.0.0.1:12001"])
	}

	if nodes["127.0.0.1:12002"].Port != 12002 {
		t.Error("Bad Result parsing , Unexpected port", nodes["127.0.0.1:12002"])
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

func TestConfigFromYml(t *testing.T) {
	cfg, err := yml.ParseYamlFile("config_test.yml")
	if err != nil {
		t.Error("Unexpected Error Parsing Yml ", err)
	}

	host, err := cfg.String("development.database.host")
	if err != nil {
		t.Error("Unexpected Error Parsing development.database.host ", err)
	}

	if host != "localhost" {
		t.Error("Unexpected result", host)
	}
}
