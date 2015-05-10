package server

type Message interface {
	MessageType() messageType
}

type messageType int

const (
	HELLO   = messageType(0)
	WELCOME = messageType(1)
	ABORT   = messageType(2)
)

// First connection message
type Hello struct {
	Id int
}

func (h *Hello) MessageType() messageType {
	return HELLO
}

// Hello Accepted
type Welcome struct {
	Id int
}

func (w *Welcome) MessageType() messageType {
	return WELCOME
}

// Hello Rejected
type Abort struct {
	Id int
}

func (a *Abort) MessageType() messageType {
	return ABORT
}
