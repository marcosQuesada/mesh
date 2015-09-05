package cluster

import (
	"github.com/marcosQuesada/mesh/message"
	"testing"
)

func TestBasicFlowOnCoordinator(t *testing.T) {
	c := Start()
	c.Send(&message.Hello{Id: message.NewId()})
	c.Send(&message.Welcome{Id: message.NewId()})
	c.Send(&message.Abort{Id: message.NewId()})
}
