package server

type Router interface {
	Route()
	Accept(p *Peer)
}

type defaultRouter struct{}

func (r *defaultRouter) Accept(p *Peer) {

}

func (r *defaultRouter) Route() {

}

/*
	switch msg.(type) {
	case *Hello:
		fmt.Println("Match", msg.(*Hello).Details["foo"])
	}

	defer close(p.exit)
	//var first bool = true //send ID on first message

	for {
		select {
		case m := <-p.rcvChan:
			msg := *m
			switch msg.MessageType() {
			case HELLO:
				fmt.Printf("Peer Hello received", msg.(*Hello))
			case WELCOME:
				fmt.Printf("Peer Welcome received", msg.(*Welcome))
			case ABORT:
				fmt.Printf("Peer Abort received", msg.(*Abort))
			default:
				fmt.Printf("unexpected type %T", msg)
			}
		case <-p.ticker.C:
			if !p.Connected() {
				fmt.Println("Connecting to ", p.remote)
				p.Connect(p.remote)
				defer p.Exit()
			}*/
