package cluster

import (
	"bufio"
	"fmt"
	"net"
)

type Client struct {
	Id          string
	inCh, OutCh chan string
	conn        net.Conn
	Finish      bool
}

func (c *Client) StartClient(id string, remote string) (inCh, outCh chan string) {
	c.inCh = make(chan string)
	c.OutCh = make(chan string)

	go func() {
		var err error
		c.conn, err = net.Dial("tcp", remote)
		if err != nil {
			fmt.Println("Error starting socket client to: ", remote, "err: ", err)
			return
		}

		fmt.Println("Client ", c.Id, "connected")

		for !c.Finish {
			go c.receiveMessages()
			message, _ := bufio.NewReader(c.conn).ReadString('\n')
			fmt.Print("Client", c.Id, " says: Received from server: "+message)
			c.OutCh <- message
		}

		defer c.conn.Close()
		defer close(c.inCh)
		defer close(c.OutCh)
	}()

	return
}

func (c *Client) Terminate() {
	c.Finish = true
}

func (c *Client) receiveMessages() {
	for {
		select {
		case m := <-c.inCh:
			fmt.Fprintf(c.conn, c.Id+" - "+m+"\n")
		default:
		}
	}
}

func (c *Client) Send(msg string) {
	c.inCh <- msg
}
