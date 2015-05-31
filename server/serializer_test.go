package server

import (
	"fmt"
	"testing"
)

func TestJsonSerializer(t *testing.T) {
	node := Node{host: "localhost", port: 13123}

	msg := Hello{
		Id:      10,
		From:    node,
		Details: map[string]interface{}{"foo": "bar"},
	}

	s := &JsonSerializer{}
	data, err := s.Serialize(msg)
	if err != nil {
		t.Error("Unexpected error serializing ", err)
	}
	fmt.Println("Data is ", msg, string(data))

	var rcvMessage Message
	rcvMessage, err = s.Deserialize(data)
	if err != nil {
		t.Error("Unexpected error deserializing ", err)
	}

	switch rcvMessage.(type) {
	case *Hello:
		h := rcvMessage.(*Hello)
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
