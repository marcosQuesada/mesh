package watch

import (
	"fmt"
	"log"
	"time"
	"reflect"
	"sync"

	"github.com/marcosQuesada/mesh/message"
)

const TIMEOUT = time.Second * 1

type RequestListener struct {
	listeners map[message.ID]chan message.Message
	timeout   time.Duration
	mutex     sync.Mutex
}

func NewRequestListener() *RequestListener {
	return &RequestListener{
		listeners: make(map[message.ID]chan message.Message, 0),
		timeout:   TIMEOUT,
		mutex:     sync.Mutex{},
	}
}

func (r *RequestListener) Notify(msg message.Message, requestID message.ID) {
	if l, ok := r.listeners[requestID]; ok {
		l <- msg
		return
	}
	log.Println("No listener found for request", requestID, "type", reflect.TypeOf(msg).String())
}

func (r *RequestListener) RegisterChan(requestID message.ID, ch chan message.Message) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.listeners[requestID] = ch
}

func (r *RequestListener) Register(requestID message.ID) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.listeners[requestID] = make(chan message.Message, 1)
}

func (r *RequestListener) Transaction(requestId message.ID) message.Message{
		r.Register(requestId)
		msg, err := r.Wait(requestId)
		if err != nil {
			log.Println("Transaction RequestListener error", requestId, err)

			msg = &message.Error{Id:requestId}
		}

	return msg
}

func (r *RequestListener) Wait(requestID message.ID) (msg message.Message, err error) {
	timeout := time.NewTimer(r.timeout)
	waitChannel, ok := r.listeners[requestID]
	if !ok {
		return nil, fmt.Errorf("unknown listener ID: %v", requestID)
	} else {
		select {
		case msg = <-waitChannel:
			timeout.Stop()

		case <-timeout.C:
			err = fmt.Errorf("timeout while waiting for message %s", requestID)
		}
	}

	close(waitChannel)
	delete(r.listeners, requestID)

	return
}
