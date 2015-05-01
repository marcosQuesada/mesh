package test

import (
	"fmt"
	"github.com/marcosQuesada/mesh"
	"testing"
	"time"
)

func TestMain(t *testing.T) {
	server := &mesh.Server{Port: 5000}
	<-server.StartServer()

	peer1 := &mesh.Client{Id: "1"}
	peer1.StartClient("1", "127.0.0.1:5000")

	peer2 := &mesh.Client{Id: "2"}
	peer2.StartClient("2", "127.0.0.1:5000")

	peer3 := &mesh.Client{Id: "3"}
	peer3.StartClient("3", "127.0.0.1:5000")

	/*	peer1.Send("Hi from 1 \n")
		peer2.Send("Hi from 2 \n")
		peer3.Send("Hi from 3 \n")
	*/
	go readMessages(peer1)
	go readMessages(peer2)
	go readMessages(peer3)

	server.Broadcast()

	time.Sleep(5 * time.Second)
	t.Fail()
}

func readMessages(peer *mesh.Client) {
	for {
		select {
		case m := <-peer.OutCh:
			fmt.Println("readMessages Received :", m)
			peer.Send("Answer from " + peer.Id + " " + m)
		default:
		}
	}
}
