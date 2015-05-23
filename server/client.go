package server

import (
	"fmt"
	"net"
)

const (
	ClientStatusConnected = Status("connected")
	ClientStatusError     = Status("error")
)

type PeerClient interface {
	Node() *Node
	Run()
	Exit()
	ReceiveChan() chan Message
	Send(Message) error
}

type Client struct {
	Peer
	node     *Node
	message  chan Message
	exitChan chan bool
}

func StartDialClient(node *Node) *Client {
	var conn net.Conn
	var err error
	for {
		conn, err = net.Dial("tcp", string(node.String()))
		if err != nil {
			fmt.Println("dial error:", err)

		} else {
			break
		}
	}
	return &Client{
		Peer:     NewJSONSocketPeer(conn),
		node:     node,
		message:  make(chan Message, 0),
		exitChan: make(chan bool),
	}
}

func (c *Client) ReceiveChan() chan Message {
	return c.message
}

func (c *Client) Run() {
	r := make(chan interface{}, 0)
	for {
		go func() {
			m, err := c.Receive()
			if err != nil {
				r <- err
			}
			r <- m
		}()

		select {
		case msg := <-r:
			switch t := msg.(type) {
			case error:
				fmt.Println("Error Receiving on server, err ", t)
			case Message:
				fmt.Println("Client Message received: ", t.(Message))
				c.message <- t.(Message)
			default:
				fmt.Printf("unexpected type %T", t)
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

func (c *Client) Node() *Node {
	return c.node
}
