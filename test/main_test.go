package test

import (
	"fmt"
	"github.com/marcosQuesada/mesh"
	"testing"
	"time"
)

func TestMain(t *testing.T) {
	mesh.StartServer(5000) /*
		go mesh.StartServer(5001)
		go mesh.StartServer(5002)*/
	time.Sleep(1 * time.Second)
	in1, out1 := mesh.StartClient("1", "127.0.0.1:5000")
	in2, out2 := mesh.StartClient("2", "127.0.0.1:5000")
	in3, out3 := mesh.StartClient("3", "127.0.0.1:5000")

	go readMessages(in1, out1)
	go readMessages(in2, out2)
	go readMessages(in3, out3)
	fmt.Println("here it goes")
	time.Sleep(3 * time.Second)
	t.Fail()
}

func sendMessages(ch chan string) {
	for {
		time.Sleep(1 * time.Second)
		ch <- "Message sended"
	}
}

func readMessages(inCh, outCh chan string) {
	for {
		select {
		case m := <-outCh:
			fmt.Println("readMessages Received :", m)
			inCh <- "Answer: " + m
		default:
		}
	}
}
