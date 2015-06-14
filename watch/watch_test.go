package watch

import (
	"fmt"
	"github.com/marcosQuesada/mesh/dispatcher"
	"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/node"
	"github.com/marcosQuesada/mesh/peer"
	"net"
	//"reflect"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	os.Exit(m.Run())
}

func TestBasicPingPongOverPipesChannel(t *testing.T) {
	nodeA := node.Node{Host: "testA", Port: 1}
	nodeB := node.Node{Host: "testB", Port: 2}
	a, b := net.Pipe()

	c1 := peer.NewAcceptor(a, nodeA)
	c1.Identify(nodeB)
	c1.Run()

	c2 := peer.NewAcceptor(b, nodeB)
	c2.Identify(nodeA)
	c2.Run()

	var wg sync.WaitGroup

	total := 5
	last := 0
	doneChan := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case r := <-c2.PingChan():
				ping := r.(*message.Ping)
				pong := &message.Pong{Id: ping.Id, From: ping.To, To: ping.From}
				c2.Send(pong)
				last = pong.Id
				if last == total {
					return
				}
			case <-doneChan:
				return
			}
		}
		return
	}()

	evCh := make(chan dispatcher.Event, 0)
	defer close(evCh)

	w := New(evCh, 1)
	fmt.Println("C1 from ", c1.From(), c1.Node())
	go w.Watch(c1)

	wg.Wait()
	close(doneChan)

	if last != total {
		t.Error("Unexpected last sample", last, "as total ", total)
	}

	w.Exit()
	c1.Exit()
	c2.Exit()
}

func TestBasicPingPongOverMultiplePipesChannel(t *testing.T) {
	nodeA := node.Node{Host: "A", Port: 1}
	nodeB := node.Node{Host: "B", Port: 2}
	nodeC := node.Node{Host: "C", Port: 3}
	nodeD := node.Node{Host: "D", Port: 4}
	nodeE := node.Node{Host: "E", Port: 5}
	nodeF := node.Node{Host: "F", Port: 6}

	a, b := net.Pipe()

	c1 := peer.NewAcceptor(a, nodeA)
	c1.Identify(nodeB)
	c1.Run()

	c1Mirror := peer.NewAcceptor(b, nodeB)
	c1Mirror.Identify(nodeA)
	c1Mirror.Run()

	c, d := net.Pipe()
	c2 := peer.NewAcceptor(c, nodeC)
	c2.Identify(nodeD)
	c2.Run()

	c2Mirror := peer.NewAcceptor(d, nodeD)
	c2Mirror.Identify(nodeC)
	c2Mirror.Run()

	e, f := net.Pipe()
	c3 := peer.NewAcceptor(e, nodeE)
	c3.Identify(nodeF)
	c3.Run()

	c3Mirror := peer.NewAcceptor(f, nodeF)
	c3Mirror.Identify(nodeE)
	c3Mirror.Run()

	var wg sync.WaitGroup

	total := 5
	last := 0
	doneChan := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case r := <-c1Mirror.PingChan():
				ping := r.(*message.Ping)
				fmt.Println("PING received", ping)
				pong := &message.Pong{Id: ping.Id, From: ping.To, To: ping.From}
				c1Mirror.Send(pong)
			case r := <-c1Mirror.PongChan():
				pong := r.(*message.Pong)
				fmt.Println("PONG received", pong)

			case r := <-c2Mirror.PingChan():
				ping := r.(*message.Ping)
				fmt.Println("PING received", ping)
				pong := &message.Pong{Id: ping.Id, From: ping.To, To: ping.From}
				c2Mirror.Send(pong)

			case r := <-c2Mirror.PongChan():
				pong := r.(*message.Pong)
				fmt.Println("PONG ", pong)

			case r := <-c3Mirror.PingChan():
				ping := r.(*message.Ping)
				pong := &message.Pong{Id: ping.Id, From: ping.To, To: ping.From}
				c3Mirror.Send(pong)
				last = pong.Id
				if last == total {
					fmt.Println("Ended")
					return
				}
			case r := <-c3Mirror.PongChan():
				pong := r.(*message.Pong)
				fmt.Println("PONG ", pong)
			case <-doneChan:
				return
			default:
			}
		}
		return
	}()

	evCh := make(chan dispatcher.Event, 0)
	defer close(evCh)

	w := New(evCh, 1)
	go w.Watch(c1)
	go w.Watch(c2)
	go w.Watch(c3)

	time.Sleep(time.Second * 2)
	c1Mirror.Send(&message.Ping{Id: 0, From: c1.From(), To: c1.Node()})
	c2Mirror.Send(&message.Ping{Id: 0, From: c2.From(), To: c2.Node()})
	c3Mirror.Send(&message.Ping{Id: 0, From: c3.From(), To: c3.Node()})

	wg.Wait()
	close(doneChan)

	if last != total {
		t.Error("Unexpected last sample", last, "as total ", total)
	}

	w.Exit()
	c1.Exit()
	c2.Exit()
	c3.Exit()
}
