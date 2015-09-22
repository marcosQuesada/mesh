package raft

import(
	"testing"
	"os"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestBasicFmtImplementation(t *testing.T) {
	f := New()
	f.Run()
}

