package dispatcher

type EventType string

type Event interface {
	GetEventType() EventType
}
