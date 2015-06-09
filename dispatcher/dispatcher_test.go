package dispatcher

import (
	"fmt"
	"testing"
	"time"
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

func TestDispatcherRun(t *testing.T) {
	l := &fakeListener{}

	d := New()
	d.RegisterListener(&OnFakeEvent{}, l.Listener)
	d.Run()

	event := &OnFakeEvent{Id: 123}
	d.EventChan <- event
	if event.Result != "Listener Called" {
		t.Error("Result Not modified")
	}

	e := make(chan Event, 0)
	d.Aggregate(e)
	eventB := &OnFakeEvent{Id: 456}
	e <- eventB

	//channel agregation needs adds minimum delay
	time.Sleep(time.Millisecond * 10)

	if eventB.Result != "Listener Called" {
		t.Error("Result Not modified")
	}

	close(d.EventChan)
	close(e)

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
	fmt.Println("Listener called")
	realEvent := e.(*OnFakeEvent)
	realEvent.Result = "Listener Called"
}
