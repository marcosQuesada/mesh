package persistence

import (
	"github.com/marcosQuesada/mesh/pkg/cluster/command"
	"os"
	"testing"
)

var persister *boltDB

func TestMain(m *testing.M) {
	persister = NewBoltDb("test.db", "foo")
	os.Exit(m.Run())
}

func TestSet(t *testing.T) {
	args := command.Args{"foo", "bar"}
	err := persister.Set(args)
	if err != nil {
		t.Error("Error On Set", err)
	}
}

func TestGet(t *testing.T) {
	args := command.Args{"foo"}
	result, err := persister.Get(args)
	if err != nil {
		t.Error("Error On Set", err)
	}

	if result != "bar" {
		t.Error("Error On Set, response don't match")
	}
}
