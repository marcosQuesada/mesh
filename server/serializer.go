package server

import (
	"encoding/json"
	//"fmt"
	"github.com/mitchellh/mapstructure"
	"log"
)

type Serializer interface {
	Serialize(Message) ([]byte, error)
	Deserialize([]byte) (Message, error)
}

type nopSerializer struct{}

func (s *nopSerializer) Serialize(m Message) []byte {
	return nil
}

func (s *nopSerializer) Deserialize(m []byte) Message {
	return nil
}

type JsonSerializer struct{}

type payload struct {
	msgType messageType
	msg     Message
}

func (s *JsonSerializer) Serialize(m Message) ([]byte, error) {
	data := map[string]interface{}{"type": m.MessageType(), "msg": m}
	log.Println("Message Serialize json ", m)
	//msg := &payload{m.MessageType(), m}
	return json.Marshal(&data)
}

func (s *JsonSerializer) Deserialize(m []byte) (Message, error) {
	payload := map[string]interface{}{}
	//p := &payload{}
	f := m
	err := json.Unmarshal(m, &payload)
	log.Println("Message json ", string(f))

	mt := messageType(int(payload["type"].(float64)))
	msg := mt.New()
	//msg.From = Node{}
	log.Println("Payload is ", payload["msg"], msg)
	err = mapstructure.Decode(payload["msg"], msg)

	return msg, err
}
