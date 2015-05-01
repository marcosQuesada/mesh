package mesh

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type Server struct {
	Port int
}

func (s *Server) StartServer() (done chan bool) {
	done = make(chan bool)
	fmt.Println("Launching server...")
	// listen on all interfaces
	go func() {
		done <- true
		ln, _ := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))

		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println("Error Accepting")
			}
			go s.handleSession(conn)
		}

	}()

	return
}

func (s *Server) handleSession(conn net.Conn) {
	for {
		// will listen for message to process ending in newline (\n)
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Print("Error Receiving:", err)
			break
		}
		// output message received
		fmt.Print("Server Message Received:", string(message))
		// sample process for string received
		newmessage := strings.ToUpper(message)
		fmt.Print("Server Says: Answer newmessage", newmessage)
		// send new string back to client
		conn.Write([]byte(newmessage + "\n"))
	}
}
