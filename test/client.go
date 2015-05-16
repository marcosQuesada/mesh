package main

import (
	"fmt"
	"github.com/marcosQuesada/mesh/server"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:12000")
	if err != nil {
		fmt.Println("dial error:", err)
		return
	}
	defer conn.Close()

	peerA := server.NewSocketPeer(conn)

	msg := server.Hello{
		Id:      10,
		Details: map[string]interface{}{"foo": "bar"},
	}
	peerA.Send(msg)

	m, err := peerA.Receive()
	fmt.Println("Message is ", m, "err", err)
}
