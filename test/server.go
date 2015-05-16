package main

import (
	"fmt"
	"github.com/marcosQuesada/mesh/server"
	//"io"
	"net"
)

type Hello struct {
	id      int
	payload map[string]interface{}
}

func main() {
	fmt.Println("Hi")

	port := ":8000"
	fmt.Println("Starting server: ", port)
	listener, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("Error starting Socket Server: ", err)
		return
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err)
			continue
		}
		peer := server.NewSocketPeer(conn)
		go handleConnection(peer)
	}
}

func handleConnection(peer *server.SocketPeer) {
	fmt.Println("Client %v connected.", peer.Conn.RemoteAddr())
	for {
		m, err := peer.Receive()
		if err != nil {
			fmt.Println("Error Receiving: ", err)
			peer.Conn.Close()
			break
		}
		err = peer.Send(m)
		if err != nil {

		}
	}

	fmt.Println("Connection from %v closed.", peer.Conn.RemoteAddr())
}
