package dispatcher

import (
	"log"
	"sync"
)

type EventType string

type Event interface {
	GetEventType() EventType
}

type Listener func(Event)

type Dispatcher interface {
	RegisterListener(Event, Listener)
	Dispatch(Event)
	Aggregate(chan Event)
	SndChan() chan Event
	Exit()
}

type defaultDispatcher struct {
	listeners map[EventType][]Listener
	mutex     sync.Mutex
	EventChan chan Event
}

func New() *defaultDispatcher {
	return &defaultDispatcher{
		listeners: make(map[EventType][]Listener, 0),
		EventChan: make(chan Event, 10),
	}
}

func (d *defaultDispatcher) RegisterListener(e Event, l Listener) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if _, ok := d.listeners[e.GetEventType()]; !ok {
		d.listeners[e.GetEventType()] = make([]Listener, 0)
	}

	d.listeners[e.GetEventType()] = append(d.listeners[e.GetEventType()], l)
}

func (d *defaultDispatcher) ConsumeEventChan() {
	for {
		select {
		case e, open := <-d.EventChan:
			if !open {
				log.Println("Closed EventChan, Exiting ConsumeEventChan Run loop")
				return
			}
			d.dispatch(e)
		}
	}
}

func (d *defaultDispatcher) Dispatch(e Event) {
	d.EventChan <- e
}

func (d *defaultDispatcher) SndChan() chan Event {
	return d.EventChan
}

//Enable event channel aggregation
func (d *defaultDispatcher) Aggregate(e chan Event) {
	go func() {
		for {
			d.EventChan <- <-e
		}
	}()
}

func (d *defaultDispatcher) Exit() {
	close(d.EventChan)
}

func (d *defaultDispatcher) dispatch(e Event) {
	if _, ok := d.listeners[e.GetEventType()]; !ok {
		return
	}

	for _, v := range d.listeners[e.GetEventType()] {
		v(e)
	}
}
