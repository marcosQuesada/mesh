package cluster

type Message interface {
	MessageType() int
}
