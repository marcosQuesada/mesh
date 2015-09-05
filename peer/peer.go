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
	PeerStatusStarted      = message.Status("started")
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
	Send(message.Message) error
	Commit(message.Message)
	State(message.Status)
	SayHello() (message.ID, error) // Pending to remove, must be internal
}

type Peer struct {
	Link
	from        n.Node
	to          n.Node
	dataChan    chan message.Message
	messageChan chan message.Message
	sendChan    chan message.Message
	exitChan    chan bool
	doneChan    chan bool
	mode        string
	state       message.Status
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
		sendChan:    make(chan message.Message, 0),
		exitChan:    make(chan bool, 2),
		doneChan:    make(chan bool, 1),
		mode:        "client",
		state:       PeerStatusStarted,
	}
}

func NewAcceptor(conn net.Conn, server n.Node) *Peer {
	return &Peer{
		Link:        NewJSONSocketLink(conn),
		from:        server,
		dataChan:    make(chan message.Message, 10),
		messageChan: make(chan message.Message, 10),
		sendChan:    make(chan message.Message, 0),
		exitChan:    make(chan bool, 1),
		doneChan:    make(chan bool, 1),
		mode:        "server",
		state:       PeerStatusStarted,
	}
}

func InitDialClient(from n.Node, destination n.Node) (*Peer, message.ID) {
	//Blocking call, wait until connection success
	p := NewDialer(from, destination)
	go p.Run()
	log.Println("Connected Dial Client from Node ", from.String(), "destination: ", destination.String(), p.Id())

	//Say Hello and wait response
	id, err := p.SayHello()
	if err != nil {
		log.Println("Error getting Hello Id ", err)
	}
	log.Println("HEllo Sended destination: ", destination.String(), id)

	return p, id
}

func (p *Peer) ReceiveChan() chan message.Message {
	return p.messageChan
}

func (p *Peer) SayHello() (u message.ID, err error) {
	msg := message.Hello{
		Id:      message.NewId(),
		From:    p.from,
		Details: map[string]interface{}{"foo": "bar"},
	}
	p.Commit(msg)

	return msg.Id, err
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
	//	log.Println("exit done ", p.Node(), p.mode, p.Id())
}

func (p *Peer) Node() n.Node {
	return p.to
}

func (p *Peer) State(s message.Status) {
	p.state = s
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

func (p *Peer) Commit(msg message.Message) {
	if p.sendChan != nil {
		p.sendChan <- msg
	}
}

func (p *Peer) receiveLoop() {
	defer close(p.dataChan)
	for {
		msg, err := p.Receive()
		if err != nil {
			if err != io.ErrClosedPipe && err != io.EOF {
				log.Println("Error Receiving: ", err, " exiting", p.from, p.mode)
			}
			return
		}
		p.dataChan <- msg
	}
}

func (p *Peer) handle() {
	defer close(p.doneChan)

	go p.handleSendChan()
	for {
		select {
		case msg, open := <-p.dataChan:
			if !open {
				//peer disconnected, exit
				go p.Exit()

				return
			}
			p.messageChan <- msg

		case <-p.exitChan:
			return
		}
	}
}

func (p *Peer) handleSendChan() {
	defer close(p.sendChan)
	defer p.nilSendChan()

	for {
		select {
		case msg, open := <-p.sendChan:
			if !open {
				log.Println("Send channel is closed, return", p.to, p.mode)
				return
			}
			p.Send(msg)
		case <-p.exitChan:
			return
		}
	}
}

func (p *Peer) nilSendChan() {
	go func() {
		p.sendChan = nil
	}()
}

// Nop Peer is Used on testing
type NopPeer struct {
	Host    string
	Port    int
	MsgChan chan message.Message
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

func (f *NopPeer) Exit() {
}
func (f *NopPeer) SayHello() (message.ID, error) {
	return message.ID("fake"), nil
}
func (f *NopPeer) State(s message.Status) {
}
func (f *NopPeer) Identify(n n.Node) {
}

func (f *NopPeer) Commit(msg message.Message) {
}

func (f *NopPeer) Mode() string {
	return ""
}

func (f *NopPeer) From() n.Node {
	return n.Node{Host: f.Host, Port: f.Port}
}
