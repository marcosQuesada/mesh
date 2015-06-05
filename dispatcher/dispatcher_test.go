package dispatcher

import (
	"testing"
)

func TestBasicDispatcher(t *testing.T) {
	l := &fakeListener{}

	d := New()

	d.RegisterListener(&OnFakeEvent{}, l.Listener)

	if len(d.listeners["OnFakeEvent"]) != 1 {
		t.Error("Unexpected dispatcher Size")
	}

	e := &OnFakeEvent{Id: 123}
	d.Dispatch(e)

	if e.Result != "Listener Called" {
		t.Error("Result Not modified")
	}

	d.Dispatch(&OnInexistentEvent{})
}

type OnFakeEvent struct {
	Id     int
	Result string
}

func (e *OnFakeEvent) GetEventType() EventType {
	return "OnFakeEvent"
}

type OnInexistentEvent struct {
}

func (e *OnInexistentEvent) GetEventType() EventType {
	return "OnInexistentEvent"
}

type fakeListener struct {
	foo int
}

func (f *fakeListener) Listener(e Event) {
	realEvent := e.(*OnFakeEvent)
	realEvent.Result = "Listener Called"
}
