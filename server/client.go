package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

const (
	ClientStatusConnected = Status("connected")
	ClientStatusError     = Status("error")
)

type PeerClient interface {
	Id() ID
	Identify(Node)
	Node() Node
	Run()
	Exit()
	ReceiveChan() chan Message
	Send(Message) error
	SayHello() // Pending to remove, must be internal
}

type Client struct {
	Peer
	from     Node
	node     Node
	message  chan Message
	exitChan chan bool
}

func StartDialClient(from Node, node Node) *Client {
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
	fmt.Println("Start dial client from ", from.String(), "to", node.String())
	return &Client{
		from:     from,
		Peer:     NewJSONSocketPeer(conn),
		node:     node,
		message:  make(chan Message, 0),
		exitChan: make(chan bool),
	}
}

func StartAcceptClient(conn net.Conn) *Client {
	return &Client{
		Peer: NewJSONSocketPeer(conn),
		//node:     node,
		message:  make(chan Message, 0),
		exitChan: make(chan bool),
	}
}

func (c *Client) ReceiveChan() chan Message {
	return c.message
}

func (c *Client) SayHello() {
	log.Println("Say Hello from ", c.from, " to : ", c.Node())
	var f Node = c.from
	var n Node = c.Node()
	msg := Hello{
		Id:      0,
		From:    f,
		Details: map[string]interface{}{"foo": "bar", "from": f.String(), "to": n.String()},
	}
	log.Println("Hello Message is ", msg)
	c.Send(msg)
}

func (c *Client) Run() {
	defer log.Println("Exiting Client Run")

	r := make(chan interface{}, 0)
	for {
		go func() {
			m, err := c.Receive()
			if err != nil {
				if err != io.EOF {
					log.Println("Error Receiving: ", err)
				}
				r <- err
				return
			}
			log.Println("Received Message on Client", m)
			r <- m
		}()

		select {
		case msg := <-r:
			switch t := msg.(type) {
			case error:
				log.Println("Error Receiving on server, err ", t)
				log.Println("Error Receiving exiting client Peer:", c.node.String())
				return
			case Message:
				log.Println("Client Message received: ", t.(Message).MessageType())
				if t.(Message).MessageType() == 0 {
					log.Println("Message Hello", t.(Message).(*Hello))
				}
				log.Println(" from: ", c.node.String())
				c.message <- t.(Message)
			default:
				log.Println("unexpected type %T", t)
			}
		case <-c.exitChan:
			fmt.Printf("Exiting Client Peer: ", c.node.String())
			c.message = nil
			c.Terminate()
			return
		}
	}
}

func (c *Client) Exit() {
	c.exitChan <- true
}

func (c *Client) Node() Node {
	return c.node
}

func (c *Client) Identify(n Node) {
	c.node = n
}
