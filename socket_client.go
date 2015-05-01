package mesh

import (
	"bufio"
	"fmt"
	"net"
	//"os"
	"time"
)

func StartClient(id string, remote string) {

	// connect to this socket
	conn, err := net.Dial("tcp", remote)
	if err != nil {
		fmt.Println("Error starting socket client to: ", remote, "err: ", err)
		return
	}

	fmt.Println("Client ", id, "connected")
	for {
		time.Sleep(1 * time.Second)
		fmt.Fprintf(conn, "hello from client"+id+"\n")
		// listen for reply
		message, _ := bufio.NewReader(conn).ReadString('\n')
		fmt.Print("Client", id, " says: Received from server: "+message)
	}
}
