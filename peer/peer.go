package peer

import (
	"io"
	"log"
	"net"
	"time"

	"github.com/marcosQuesada/mesh/message"
	n "github.com/marcosQuesada/mesh/node"
)

const (
	PeerStatusConnecting   = message.Status("connecting")
	PeerStatusConnected    = message.Status("connected")
	PeerStatusDisconnected = message.Status("disconnected")
	PeerStatusAbort        = message.Status("abort")
	PeerStatusError        = message.Status("error")
	PeerStatusUnknown      = message.Status("unknown")
)

type NodePeer interface {
	Id() ID
	Identify(n.Node)
	Node() n.Node
	From() n.Node
	Mode() string
	Run()
	Exit()
	ReceiveChan() chan message.Message
	PingChan() chan message.Message
	PongChan() chan message.Message
	Send(message.Message) error
	SayHello() // Pending to remove, must be internal
}

type Peer struct {
	Link
	from        n.Node
	to          n.Node
	dataChan    chan message.Message
	messageChan chan message.Message
	pingChan    chan message.Message
	pongChan    chan message.Message
	exitChan    chan bool
	doneChan    chan bool
	mode        string
}

func NewDialer(from n.Node, destination n.Node) *Peer {
	var conn net.Conn
	var err error
	for {
		conn, err = net.Dial("tcp", string(destination.String()))
		if err != nil {
			log.Println("dial error:", err)
			time.Sleep(time.Second)
		} else {
			break
		}
	}
	return &Peer{
		from:        from,
		Link:        NewJSONSocketLink(conn),
		to:          destination,
		dataChan:    make(chan message.Message, 10),
		messageChan: make(chan message.Message, 10),
		pingChan:    make(chan message.Message, 10),
		pongChan:    make(chan message.Message, 10),
		exitChan:    make(chan bool, 2),
		doneChan:    make(chan bool, 1),
		mode:        "client",
	}
}

func NewAcceptor(conn net.Conn, server n.Node) *Peer {
	return &Peer{
		Link:        NewJSONSocketLink(conn),
		from:        server,
		dataChan:    make(chan message.Message, 10),
		messageChan: make(chan message.Message, 10),
		pingChan:    make(chan message.Message, 10),
		pongChan:    make(chan message.Message, 10),
		exitChan:    make(chan bool, 1),
		doneChan:    make(chan bool, 1),
		mode:        "server",
	}
}

func (p *Peer) ReceiveChan() chan message.Message {
	return p.messageChan
}

func (p *Peer) PingChan() chan message.Message {
	return p.pingChan
}

func (p *Peer) PongChan() chan message.Message {
	return p.pongChan
}

func (p *Peer) SayHello() {
	msg := message.Hello{
		Id:      0,
		From:    p.from,
		Details: map[string]interface{}{"foo": "bar"},
	}
	p.Send(msg)
}

func (p *Peer) Run() {
	go p.handle()
	//start receive loop
	p.receiveLoop()
}

func (p *Peer) Exit() {
	close(p.exitChan)
	<-p.doneChan

	close(p.messageChan)
	log.Println("BYE ", p.From())
}

func (p *Peer) Node() n.Node {
	return p.to
}

//Used on dev Only!
func (p *Peer) From() n.Node {
	return p.from
}

func (p *Peer) Mode() string {
	return p.mode
}

func (p *Peer) Identify(n n.Node) {
	p.to = n
}

func (p *Peer) receiveLoop() {
	defer close(p.dataChan)
	for {
		msg, err := p.Receive()
		if err != nil {
			if err != io.ErrClosedPipe && err != io.EOF {
				log.Println("Error Receiving: ", err, " exiting", p.from)
			}
			return
		}
		p.dataChan <- msg
	}
}

func (p *Peer) handle() {
	defer close(p.pongChan)
	defer close(p.pingChan)
	defer close(p.doneChan)

	for {
		select {
		case msg, open := <-p.dataChan:
			if !open {
				log.Println("Data channel is closed, return", p.to)
				return
			}
			switch msg.(type) {
			/*			case *message.Ping:
							p.pingChan <- msg
							continue

						case *message.Pong:
							p.pongChan <- msg
							continue
			*/
			default:
				p.messageChan <- msg
				continue
			}

		case <-p.exitChan:
			log.Println("Peer handleLoop exits", p.to)
			return
		}
	}
}

// Nop Peer is Used on testing
type NopPeer struct {
	Host         string
	Port         int
	MsgChan      chan message.Message
	PingPongChan chan message.Message
}

func (f *NopPeer) Node() n.Node {
	return n.Node{Host: f.Host, Port: f.Port}
}
func (f *NopPeer) Id() ID {
	return ID(0)
}
func (f *NopPeer) Run() {
}
func (f *NopPeer) Send(m message.Message) error {
	f.MsgChan <- m
	return nil
}
func (f *NopPeer) ReceiveChan() chan message.Message {
	return f.MsgChan
}

func (f *NopPeer) PingChan() chan message.Message {
	return f.PingPongChan
}
func (f *NopPeer) PongChan() chan message.Message {
	return f.PingPongChan
}
func (f *NopPeer) Exit() {
}
func (f *NopPeer) SayHello() {
}
func (f *NopPeer) Identify(n n.Node) {
}

func (f *NopPeer) Mode() string {
	return ""
}

func (f *NopPeer) From() n.Node {
	return n.Node{Host: f.Host, Port: f.Port}
}
