package serializer

import (
	m "github.com/marcosQuesada/mesh/message"
	n "github.com/marcosQuesada/mesh/node"
	"testing"
)

func TestJsonSerializer(t *testing.T) {
	node := n.Node{Host: "localhost", Port: 13123}

	msg := m.Hello{
		Id:      m.NewId(),
		From:    node,
		Details: map[string]interface{}{"foo": "bar"},
	}

	s := &JsonSerializer{}
	data, err := s.Serialize(msg)
	if err != nil {
		t.Error("Unexpected error serializing ", err)
	}

	var rcvMessage m.Message
	rcvMessage, err = s.Deserialize(data)
	if err != nil {
		t.Error("Unexpected error deserializing ", err)
	}

	switch rcvMessage.(type) {
	case *m.Hello:
		h := rcvMessage.(*m.Hello)
		if msg.Id != h.Id {
			t.Error("Message Ids don't match", msg.Id, h.Id)
		}

		if "bar" != h.Details["foo"] {
			t.Error("Message Payload don't match")
		}
	default:
		t.Error("Wrong type")
	}
}
