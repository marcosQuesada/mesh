package watch

import (
	"fmt"
	"log"
	"time"

	"sync"

	"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/node"
)

type requestListener struct {
	listeners map[string]chan message.Message
	timeout   time.Duration
	mutex     sync.Mutex
}

func (r *requestListener) Id(n node.Node, id int) string {
	return fmt.Sprintf("node-%s-id-%d", n.String(), id)
}

func (r *requestListener) notify(msg message.Message, requestID string) {
	log.Println("Notify", requestID, "type", msg)
	if l, ok := r.listeners[requestID]; ok {
		l <- msg
		log.Println("Notify done")
		return
	}
	log.Println("No listener found for request", requestID, "type", msg.MessageType())

}

func (r *requestListener) register(requestID string) {
	log.Println("Register", requestID)
	wait := make(chan message.Message, 1)
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.listeners[requestID] = wait
}

func (r *requestListener) wait(requestID string) (msg message.Message, err error) {
	timeout := time.NewTimer(time.Second * 2)
	wait, ok := r.listeners[requestID]
	if !ok {
		return nil, fmt.Errorf("unknown listener ID: %v", requestID)
	} else {
		select {
		case msg = <-wait:
			log.Println("Wait done", requestID)
			timeout.Stop()

		case <-timeout.C: //time.After(r.timeout)
			err = fmt.Errorf("timeout while waiting for message")
			panic(err)
		}
	}
	close(wait)
	delete(r.listeners, requestID)

	return
}
