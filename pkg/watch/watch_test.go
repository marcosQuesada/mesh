package watch

/*func TestBasicPingPongOverPipesChannel(t *testing.T) {
	total := 2
	timeInterval := 1
	nodeA := node.Node{Host: "A", Port: 1}
	nodeB := node.Node{Host: "B", Port: 2}
	a, b := net.Pipe()

	c1 := peer.NewAcceptor(a, nodeA)
	c1.Identify(nodeB)
	go c1.Run()

	c1Mirror := peer.NewAcceptor(b, nodeB)
	c1Mirror.Identify(nodeA)
	go c1Mirror.Run()

	var wg sync.WaitGroup

	last := 0
	doneChan := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case r, open := <-c1Mirror.ReceiveChan():
				if !open {
					fmt.Println("ReceiveChan Closed")
					return
				}
				ping := r.(*message.Ping)
				pong := &message.Pong{Id: ping.Id, From: ping.To, To: ping.From}
				c1Mirror.Send(pong)
				last = pong.Id
				if last == total {
					fmt.Println("Exiting")

					return
				}
			case <-doneChan:
				fmt.Println("Done")
				return
			}
		}
		return
	}()
	evCh := make(chan dispatcher.Event, 0)
	defer close(evCh)

	w := New(evCh, timeInterval)
	fmt.Println("C1 from ", c1.From(), c1.Node())
	go w.Watch(c1)

	wg.Wait()
	close(doneChan)

	if last != total {
		t.Error("Unexpected last sample", last, "as total ", total)
	}

	w.Exit()

	fmt.Println("1 exit")
	c1.Exit()
	fmt.Println("C1 Exit")
	c1Mirror.Exit()
	fmt.Println("C1 mirror exit done")
}

func TestPingPongOverMultiplePipesChannel(t *testing.T) {
	total := 3
	timeInterval := 1
	nodeA := node.Node{Host: "c1", Port: 1}
	nodeB := node.Node{Host: "c1mirror", Port: 2}
	nodeC := node.Node{Host: "c2", Port: 3}
	nodeD := node.Node{Host: "c2mirror", Port: 4}
	nodeE := node.Node{Host: "c3", Port: 5}
	nodeF := node.Node{Host: "c3mirror", Port: 6}

	a, b := net.Pipe()

	c1 := peer.NewAcceptor(a, nodeA)
	c1.Identify(nodeB)
	go c1.Run()

	c1Mirror := peer.NewAcceptor(b, nodeB)
	c1Mirror.Identify(nodeA)
	go c1Mirror.Run()

	c, d := net.Pipe()
	c2 := peer.NewAcceptor(c, nodeC)
	c2.Identify(nodeD)
	go c2.Run()

	c2Mirror := peer.NewAcceptor(d, nodeD)
	c2Mirror.Identify(nodeC)
	go c2Mirror.Run()

	e, f := net.Pipe()
	c3 := peer.NewAcceptor(e, nodeE)
	c3.Identify(nodeF)
	go c3.Run()

	c3Mirror := peer.NewAcceptor(f, nodeF)
	c3Mirror.Identify(nodeE)
	go c3Mirror.Run()

	var wg sync.WaitGroup

	last := 0
	doneChan := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case r, open := <-c1Mirror.ReceiveChan():
				if !open {
					fmt.Println("Pingchan closed , exits")
					return
				}
				ping, ok := r.(*message.Ping)
				if !ok {
					pong := r.(*message.Pong)
					fmt.Println("c1Mirror PONG received from ", pong.From.String(), "ID ", pong.Id)
					continue
				}
				fmt.Println("c1Mirror PING received from ", ping.From.String())
				pong := &message.Pong{Id: ping.Id, From: ping.To, To: ping.From}
				c1Mirror.Send(pong)

			case r, open := <-c2Mirror.ReceiveChan():
				if !open {
					fmt.Println("Pingchan closed , exits")
					return
				}
				ping, ok := r.(*message.Ping)
				if !ok {
					pong := r.(*message.Pong)
					fmt.Println("c1Mirror PONG received from ", pong.From.String(), "ID ", pong.Id)
					continue
				}
				pong := &message.Pong{Id: ping.Id, From: ping.To, To: ping.From}
				c2Mirror.Send(pong)

			case r, open := <-c3Mirror.ReceiveChan():
				if !open {
					fmt.Println("PingChan Closed, exits")
					return
				}
				ping, ok := r.(*message.Ping)
				if !ok {
					pong := r.(*message.Pong)
					fmt.Println("c1Mirror PONG received from ", pong.From.String(), "ID ", pong.Id)
					continue
				}
				fmt.Println("c3Mirror PING received from ", ping.From.String())
				pong := &message.Pong{Id: ping.Id, From: ping.To, To: ping.From}
				c3Mirror.Send(pong)
				last = pong.Id
				if last == total {
					return
				}

			case <-doneChan:
				return
			default:
			}
		}
		return
	}()

	evCh := make(chan dispatcher.Event, 0)

	w := New(evCh, timeInterval)
	go w.Watch(c1)
	go w.Watch(c2)
	go w.Watch(c3)

	time.Sleep(time.Second * 3)
	c1Mirror.Commit(&message.Ping{Id: 0, From: c1Mirror.From(), To: c1Mirror.Node()})
	c2Mirror.Commit(&message.Ping{Id: 0, From: c2Mirror.From(), To: c2Mirror.Node()})
	c3Mirror.Commit(&message.Ping{Id: 0, From: c3Mirror.From(), To: c3Mirror.Node()})

	wg.Wait()
	close(doneChan)

	if last != total {
		t.Error("Unexpected last sample", last, "as total ", total)
	}

	w.Exit()
	close(evCh)

	//time.Sleep(time.Second)
	c1.Exit()
	fmt.Println("1 exit")
	c2.Exit()
	fmt.Println("2 exit")
	c3.Exit()
	fmt.Println("3 exit")

	c1Mirror.Exit()
	fmt.Println("1 mirror")
	c2Mirror.Exit()
	fmt.Println("2 mirror")
	c3Mirror.Exit()
	fmt.Println("3 mirror")
}
*/
