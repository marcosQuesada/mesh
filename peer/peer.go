package peer

import (
	"github.com/marcosQuesada/mesh/message"
	n "github.com/marcosQuesada/mesh/node"
	"io"
	"log"
	"net"
	"time"
)

const (
	PeerStatusConnecting = message.Status("connecting")
	PeerStatusConnected  = message.Status("connected")
	PeerStatusAbort      = message.Status("abort")
	PeerStatusError      = message.Status("error")
	PeerStatusUnknown    = message.Status("unknown")
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
	Send(message.Message) error
	SayHello() // Pending to remove, must be internal
}

type Peer struct {
	Link
	from        n.Node
	node        n.Node
	messageChan chan message.Message
	pingChan    chan message.Message
	exitChan    chan bool
	mode        string
}

func NewDialer(from n.Node, node n.Node) *Peer {
	var conn net.Conn
	var err error
	for {
		conn, err = net.Dial("tcp", string(node.String()))
		if err != nil {
			log.Println("dial error:", err)
			time.Sleep(time.Second)
		} else {
			break
		}
	}
	//fmt.Println("Start dial Peer from ", from.String(), "to", node.String())
	return &Peer{
		from:        from,
		Link:        NewJSONSocketLink(conn),
		node:        node,
		messageChan: make(chan message.Message, 0),
		exitChan:    make(chan bool),
		mode:        "client",
	}
}

func NewAcceptor(conn net.Conn, node n.Node) *Peer {
	return &Peer{
		Link:        NewJSONSocketLink(conn),
		node:        node,
		messageChan: make(chan message.Message, 0),
		exitChan:    make(chan bool),
		mode:        "server",
	}
}

func (p *Peer) ReceiveChan() chan message.Message {
	return p.messageChan
}

func (p *Peer) PingChan() chan message.Message {
	return p.pingChan
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
	//defer log.Println("Exiting Peer Link type ", p.mode, "-", p.node.String())

	response := make(chan interface{}, 0)
	done := make(chan bool)
	go func(exit chan bool) {
		select {
		case <-exit:
			return
		default:
			for {
				m, err := p.Receive()
				if err != nil {
					if err != io.EOF {
						log.Println("Error Receiving: ", err, " exiting")
					}
					response <- err
					return
				}
				response <- m
			}
		}
	}(done)

	//Required to allow message chan readers
	go func() {
		for {
			select {
			case msg := <-response:
				switch t := msg.(type) {
				case error:
					log.Println("Error Receiving on server, err ", t, "exiting Peer Link:", p.node.String())
					p.exitChan <- true
					return
				case message.Message:
					m := t.(message.Message)
					switch m.(type) {
					case *message.Ping:
						log.Println("Received Ping")
						p.pingChan <- m
						continue
					case *message.Pong:
						log.Println("Received Pong")
						p.pingChan <- m
						continue
					default:
						p.messageChan <- t.(message.Message)
					}
				default:
					log.Println("unexpected type %T", t)
				}
			case <-p.exitChan:
				done <- true
				close(p.messageChan)
				p.messageChan = nil
				p.Terminate()

				return
			}
		}
	}()
}

func (p *Peer) Exit() {
	p.exitChan <- true
}

func (p *Peer) Node() n.Node {
	return p.node
}

//Used on dev Only!
func (p *Peer) From() n.Node {
	return p.from
}

func (p *Peer) Mode() string {
	return p.mode
}

func (p *Peer) Identify(n n.Node) {
	p.node = n
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
