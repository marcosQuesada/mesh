package server

import (
	"testing"
)

func TestBasicFlowOnCoordinator(t *testing.T) {
	c := Start()
	c.Send(&Hello{Id: 11})
	c.Send(&Welcome{Id: 11})
	c.Send(&Abort{Id: 11})
}
