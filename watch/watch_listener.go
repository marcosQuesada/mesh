package watch

import (
	"fmt"
	"log"
	"time"

	"sync"

	"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/node"
)

const TIMEOUT = time.Second * 2

type requestListener struct {
	listeners map[string]chan message.Message
	timeout   time.Duration
	mutex     sync.Mutex
}

func newRequestListener() *requestListener {
	return &requestListener{
		listeners: make(map[string]chan message.Message, 0),
		timeout:   TIMEOUT,
	}
}

func (r *requestListener) Id(n node.Node, id int) string {
	return fmt.Sprintf("node-%s-id-%d", n.String(), id)
}

func (r *requestListener) notify(msg message.Message, requestID string) {
	if l, ok := r.listeners[requestID]; ok {
		l <- msg
		log.Println("Notify done", requestID)
		return
	}
	log.Println("No listener found for request", requestID, "type", msg.MessageType())
}

func (r *requestListener) register(requestID string) {
	log.Println("Register", requestID)
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.listeners[requestID] = make(chan message.Message, 1)
}

func (r *requestListener) wait(requestID string) (msg message.Message, err error) {
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
