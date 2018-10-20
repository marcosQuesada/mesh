package serializer

import (
	"encoding/json"
	"github.com/marcosQuesada/mesh/pkg/message"
	"github.com/mitchellh/mapstructure"
)

type Serializer interface {
	Serialize(message.Message) ([]byte, error)
	Deserialize([]byte) (message.Message, error)
}

type nopSerializer struct{}

func (s *nopSerializer) Serialize(m message.Message) []byte {
	return nil
}

func (s *nopSerializer) Deserialize(m []byte) message.Message {
	return nil
}

type JsonSerializer struct{}

func (s *JsonSerializer) Serialize(m message.Message) ([]byte, error) {
	t := m.MessageType()

	data := map[string]interface{}{"type": t, "msg": m}
	return json.Marshal(&data)
}

func (s *JsonSerializer) Deserialize(data []byte) (message.Message, error) {
	payload := map[string]interface{}{}
	err := json.Unmarshal(data, &payload)

	msg := message.MsgType(int(payload["type"].(float64))).New()
	err = mapstructure.Decode(payload["msg"], msg)

	return msg, err
}
