package dispatcher

import (
	"fmt"
	"sync"
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
	d.dispatch(e)

	if result.get() != "Listener Called" {
		t.Error("Result Not modified")
	}

	d.dispatch(&OnInexistentEvent{})
}

func TestDispatcherRun(t *testing.T) {
	l := &fakeListener{}

	d := New()
	d.RegisterListener(&OnFakeEvent{}, l.Listener)
	go d.Run()

	event := &OnFakeEvent{Id: 123}
	d.EventChan <- event
	time.Sleep(time.Millisecond * 100)
	if result.get() != "Listener Called" {
		t.Error("Result Not modified")
	}

	e := make(chan Event, 0)
	go d.Aggregate(e)
	eventB := &OnFakeEvent{Id: 456}
	e <- eventB

	//channel agregation needs adds minimum delay
	time.Sleep(time.Millisecond * 10)

	if result.get() != "Listener Called" {
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
	foo   int
	mutex sync.Mutex
}

func (f *fakeListener) Listener(e Event) {
	fmt.Println("Listener called")
	//realEvent := e.(*OnFakeEvent)
	result.set("Listener Called")
}

var result fakeResult

type fakeResult struct {
	result string
	mutex  sync.Mutex
}

func (f *fakeResult) set(v string) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.result = v
}

func (f *fakeResult) get() string {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return f.result
}
