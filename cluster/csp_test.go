package cluster

import (
	"github.com/marcosQuesada/mesh/message"
	"testing"
)

func TestBasicFlowOnCoordinator(t *testing.T) {
	c := Start()
	c.Send(&message.Hello{Id: 11})
	c.Send(&message.Welcome{Id: 11})
	c.Send(&message.Abort{Id: 11})
}
