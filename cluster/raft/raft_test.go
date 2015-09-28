package raft

import(
	"testing"
	"os"
	"github.com/marcosQuesada/mesh/node"
/*	"time"
	"log"*/
)

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
	result  bool
	responses votationResult
}{
	{true , votationResult{"B", []string{"B", "B"}}},
	{false, votationResult{"B", []string{"A", "C"}}},
	{false, votationResult{"A", []string{"B", "B", "A"}}},
}

func TestEvaluateResponses(t *testing.T) {
	r := &Raft{
		node:         node.Node{"A",1},
		mates:        []node.Node{node.Node{"B",2}, node.Node{"C",3}},
	}

	for _, item := range flagtests {
		if item.result != r.evaluate(item.responses) {
			t.Error("unexpected result ", item)
		}
	}
}