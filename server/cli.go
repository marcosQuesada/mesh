package server

import (
	"bufio"
	//"fmt"
	"log"
	"net"
	"strings"
)

type CliSession struct {
	conn   net.Conn
	server *Server // Used as reflection
	finish bool
}

func (c *CliSession) handle() {
	defer c.conn.Close()
	for !c.finish {
		message, err := bufio.NewReader(c.conn).ReadString('\n')
		if err != nil {
			log.Print("Error Receiving on server, err ", err)
			return
		}

		c.process(message)

		log.Println("Server received Message ", message)
	}
}

func (c *CliSession) process(line string) {
	var response string = "Ask for HELP! \n"

	switch strings.Trim(strings.ToUpper(line), "\r\n") {
	case "HELP":
		response = "Available commands:\n LIST \n SHUTDOWN \n EXIT \n"
	case "LIST":
		/*		response = fmt.Sprintf("Total Peers: %d \n", len(c.server.Links()))
				for key, _ := range c.server.Links() {
					response = response + " " + key + "\n"
				}*/
	case "SHUTDOWN":
		c.server.Close()
		response = "Server Shutting down \n"
	case "EXIT":
		c.finish = true
		response = "Closing cli session, Bye Bye! \n"
	}

	c.send(response)
}

func (c *CliSession) send(response string) {
	_, err := c.conn.Write([]byte(response))
	if err != nil {
		log.Println("Error Writting on socket ", err)
	}
}
