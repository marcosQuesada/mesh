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

	go readMessages(peer1)
	go readMessages(peer2)
	go readMessages(peer3)

	server.Broadcast()

	time.Sleep(2 * time.Second)
	server.Terminate()

	peer1.Terminate()
	peer2.Terminate()
	peer3.Terminate()
}

func TestRing(t *testing.T) {
	time.Sleep(2 * time.Second)
	//Start all server nodes
	server := &mesh.Server{Port: 5001}
	<-server.StartServer()

	server1 := &mesh.Server{Port: 5002}
	<-server1.StartServer()

	server2 := &mesh.Server{Port: 5003}
	<-server2.StartServer()

	//Simulates Client from Server 1
	peer2 := &mesh.Client{Id: "2"}
	peer2.StartClient("2", "127.0.0.1:5002")

	//Simulates Client from Server 2
	peer3 := &mesh.Client{Id: "3"}
	peer3.StartClient("3", "127.0.0.1:5003")

	//Simulates Client from Server 3
	peer1 := &mesh.Client{Id: "1"}
	peer1.StartClient("1", "127.0.0.1:5001")

	time.Sleep(1 * time.Second)
	peer1.Send("Hi from 1 \n")

	for i := 0; i < 4; i++ {
		peer2.Send(read(peer1)) // What server 1 gets is forwarded to server 2
		peer3.Send(read(peer2)) // Forward from server 2 to 3
		peer1.Send(read(peer3)) // Completes Loop joinning server 3 with the first one
	}
}

func readMessages(peer *mesh.Client) {
	for !peer.Finish {
		select {
		case m := <-peer.OutCh:
			fmt.Println("readMessages Received :", m)
			peer.Send("Answer from " + peer.Id + " " + m)
		default:
		}
	}
	return
}

func read(peer *mesh.Client) string {
	for {
		select {
		case m := <-peer.OutCh:
			fmt.Println("readMessages Received :", m)
			return m
		default:
		}
	}
}
