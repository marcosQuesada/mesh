package dispatcher

import (
	"sync"
)

type Listener func(Event)

type Dispatcher interface {
	RegisterListener(Event, Listener)
	Dispatch(Event)
}

type DefaultDispatcher struct {
	listeners map[EventType][]Listener
	mutex     sync.Mutex
}

func (d *DefaultDispatcher) RegisterListener(e Event, l Listener) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if _, ok := d.listeners[e.GetEventType()]; !ok {
		d.listeners[e.GetEventType()] = make([]Listener, 0)
	}

	d.listeners[e.GetEventType()] = append(d.listeners[e.GetEventType()], l)
}

func (d *DefaultDispatcher) Dispatch(e Event) {
	if _, ok := d.listeners[e.GetEventType()]; !ok {
		return
	}

	for _, v := range d.listeners[e.GetEventType()] {
		v(e)
	}
}
