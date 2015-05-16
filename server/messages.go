package server

type Message interface {
	MessageType() messageType
}

type messageType int

func (mt messageType) New() Message {
	switch mt {
	case HELLO:
		return &Hello{}
	case WELCOME:
		return &Welcome{}
	case ABORT:
		return &Abort{}
	case PING:
		return &Ping{}
	case PONG:
		return &Pong{}
	case GOODBYE:
		return &GoodBye{}
	}

	return nil
}

const (
	HELLO   = messageType(0)
	WELCOME = messageType(1)
	ABORT   = messageType(2)
	PING    = messageType(3)
	PONG    = messageType(4)
	GOODBYE = messageType(5)
)

// First connection message
type Hello struct {
	Id      int
	Details map[string]interface{}
}

func (h Hello) MessageType() messageType {
	return HELLO
}

// Hello Accepted
type Welcome struct {
	Id      int
	Details map[string]interface{}
}

func (w Welcome) MessageType() messageType {
	return WELCOME
}

// Hello Rejected
type Abort struct {
	Id      int
	Details map[string]interface{}
}

func (a Abort) MessageType() messageType {
	return ABORT
}

// Ping request to a remote node
type Ping struct {
	Id      int
	Details map[string]interface{}
}

func (p Ping) MessageType() messageType {
	return PING
}

// Pong response as a ping request
type Pong struct {
	Id      int
	Details map[string]interface{}
}

func (p Pong) MessageType() messageType {
	return PONG
}

type GoodBye struct {
	Id      int
	Details map[string]interface{}
}

func (g GoodBye) MessageType() messageType {
	return GOODBYE
}
