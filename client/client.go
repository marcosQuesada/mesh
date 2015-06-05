package client

import (
	"github.com/marcosQuesada/mesh/message"
	n "github.com/marcosQuesada/mesh/node"
	"io"
	"log"
	"net"
	"time"
)

const (
	ClientStatusConnected = message.Status("connected")
	ClientStatusAbort     = message.Status("abort")
	ClientStatusError     = message.Status("error")
)

type PeerClient interface {
	Id() ID
	Identify(n.Node)
	Node() n.Node
	From() n.Node
	Mode() string
	Run()
	Exit()
	ReceiveChan() chan message.Message
	Send(message.Message) error
	SayHello() // Pending to remove, must be internal
}

type Client struct {
	Peer
	from        n.Node
	node        n.Node
	messageChan chan message.Message
	exitChan    chan bool
	mode        string
}

func StartDial(from n.Node, node n.Node) *Client {
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
	//fmt.Println("Start dial client from ", from.String(), "to", node.String())
	return &Client{
		from:        from,
		Peer:        NewJSONSocketPeer(conn),
		node:        node,
		messageChan: make(chan message.Message, 0),
		exitChan:    make(chan bool),
		mode:        "client",
	}
}

func StartAccept(conn net.Conn, node n.Node) *Client {
	return &Client{
		Peer:        NewJSONSocketPeer(conn),
		node:        node,
		messageChan: make(chan message.Message, 0),
		exitChan:    make(chan bool),
		mode:        "server",
	}
}

func (c *Client) ReceiveChan() chan message.Message {
	return c.messageChan
}

func (c *Client) SayHello() {
	msg := message.Hello{
		Id:      0,
		From:    c.from,
		Details: map[string]interface{}{"foo": "bar"},
	}
	c.Send(msg)
}

func (c *Client) Run() {
	defer log.Println("Exiting Client Peer type ", c.mode, "-", c.node.String())

	response := make(chan interface{}, 0)
	done := make(chan bool)
	go func(exit chan bool) {
		select {
		case <-exit:
			return
		default:
			for {
				m, err := c.Receive()
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
					log.Println("Error Receiving on server, err ", t, "exiting client Peer:", c.node.String())
					c.exitChan <- true
					return
				case message.Message:
					c.messageChan <- t.(message.Message)
				default:
					log.Println("unexpected type %T", t)
				}
			case <-c.exitChan:
				done <- true
				c.messageChan = nil
				c.Terminate()

				return
			}
		}
	}()
	//wait to start receive loop & forward loop to message chan
}

func (c *Client) Exit() {
	c.exitChan <- true
}

func (c *Client) Node() n.Node {
	return c.node
}

//Used on dev Only!
func (c *Client) From() n.Node {
	return c.from
}

func (c *Client) Mode() string {
	return c.mode
}

func (c *Client) Identify(n n.Node) {
	c.node = n
	log.Println("Identified Server peer as ", n.String())
}