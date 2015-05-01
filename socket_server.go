package mesh

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

func StartServer(port int) {

	fmt.Println("Launching server...")

	// listen on all interfaces
	ln, _ := net.Listen("tcp", fmt.Sprintf(":%d", port))

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error Accepting")
		}
		go handleSession(conn)
	}
}

func handleSession(conn net.Conn) {
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
