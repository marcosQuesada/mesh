package request

import (
	"fmt"
	"log"
	"reflect"
	"sync"
	"time"

	"github.com/marcosQuesada/mesh/pkg/message"
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

func (r *RequestListener) Register(requestID message.ID) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.listeners[requestID] = make(chan message.Message, 1)
}

func (r *RequestListener) Notify(msg message.Message, requestID message.ID) {
	l, ok := r.listeners[requestID]
	if !ok {
		log.Println("No listener found for request", requestID, "type", reflect.TypeOf(msg).String())
		return
	}

	l <- msg
}

func (r *RequestListener) Wait(requestID message.ID) (msg message.Message, err error) {
	waitChannel, ok := r.listeners[requestID]
	if !ok {
		return nil, fmt.Errorf("unknown listener ID: %v", requestID)
	} else {
		timeout := time.NewTimer(time.Second * 1)
		select {
		case msg = <-waitChannel:
			timeout.Stop()
		case <-timeout.C:
			err = fmt.Errorf("timeout while waiting for message %s", requestID)
		}
	}
	close(waitChannel)

	r.mutex.Lock()
	delete(r.listeners, requestID)
	r.mutex.Unlock()

	return
}

func (r *RequestListener) RegisterAndWait(requestID message.ID) (msg message.Message, err error) {
	r.Register(requestID)

	return r.Wait(requestID)
}
