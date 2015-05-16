package server

import (
	"testing"
)

func TestJsonSerializer(t *testing.T) {
	msg := Hello{
		Id:      10,
		Details: map[string]interface{}{"foo": "bar"},
	}

	s := &JsonSerializer{}
	data, err := s.Serialize(msg)
	if err != nil {
		t.Error("Unexpected error serializing ", err)
	}

	var rcvMessage Message
	rcvMessage, err = s.Deserialize(data)
	if err != nil {
		t.Error("Unexpected error deserializing ", err)
	}

	switch rcvMessage.(type) {
	case *Hello:
		h := rcvMessage.(*Hello)
		if msg.Id != h.Id {
			t.Error("Message Ids don't match")
		}

		if "bar" != h.Details["foo"] {
			t.Error("Message Payload don't match")
		}
	default:
		t.Error("Wrong type")
	}
}
