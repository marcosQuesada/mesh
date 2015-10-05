package raft

import (
	"github.com/marcosQuesada/mesh/node"
	"os"
	"testing"
	/*	"time"
		"log"*/)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

/*func TestBasicFmtImplementation(t *testing.T) {
	f := New(node.Node{},  []node.Node{node.Node{"B",2}, node.Node{"C",3}})
	fmt := &Fmt{state: f}
	go fmt.Run()
	time.Sleep(time.Second * 1)
	f.response <- "foo"
	log.Println("fired")
	time.Sleep(time.Second * 10)
	f.response <- "bar"
	log.Println("fired")
	time.Sleep(time.Second * 10)
	f.response <- "barsdad"
	time.Sleep(time.Second * 10)
}*/

var flagtests = []struct {
	result    bool
	responses PoolResult
}{
	{true, PoolResult{"B": true, "C": true}},
	{true, PoolResult{"B": true, "C": false}},
	//{false, PoolResult{"B":true, "C":false, "D":false}},
	//{true, PoolResult{"B":true, "C":false, "D":false, "E":true}},
}

func TestEvaluateResponses(t *testing.T) {
	r := &Raft{
		node:  node.Node{"A", 1},
		mates: map[string]node.Node{"B:2": node.Node{"B", 2}, "C:3": node.Node{"C", 3}},
	}

	for _, item := range flagtests {
		if item.result != r.evaluate(item.responses) {
			t.Error("unexpected result ", item)
		}
	}
}

var flagtestsBig = []struct {
	result    bool
	responses PoolResult
}{
	{true, PoolResult{"B": true, "C": true, "D": true, "E": true, "F": false}},
	{false, PoolResult{"B": true, "C": false, "D": false, "E": false, "F": false}},
	{false, PoolResult{"B": true, "C": false, "D": false, "E": true, "F": false}}, //empate!
	{true, PoolResult{"B": true, "C": false, "D": false, "E": true, "F": true}},
}

func TestEvaluateResponsesFiveNodes(t *testing.T) {
	r := &Raft{
		node:  node.Node{"A", 1},
		mates: map[string]node.Node{"B:2": node.Node{"B", 2}, "C:3": node.Node{"C", 3}, "D:4": node.Node{"D", 4}, "E:5": node.Node{"E", 5}, "F:6": node.Node{"F", 6}},
	}

	for _, item := range flagtestsBig {
		if item.result != r.evaluate(item.responses) {
			t.Error("unexpected result ", item)
		}
	}
}
