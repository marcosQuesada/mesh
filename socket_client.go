package mesh

import (
	"bufio"
	"fmt"
	"net"
)

func StartClient(id string, remote string) (inCh, outCh chan string) {
	inCh = make(chan string)
	outCh = make(chan string)
	go func() {
		conn, err := net.Dial("tcp", remote)
		if err != nil {
			fmt.Println("Error starting socket client to: ", remote, "err: ", err)
			return
		}

		fmt.Println("Client ", id, "connected")
		fmt.Fprintf(conn, "hi from "+id+"\n")
		for {
			go receiveMessages(inCh, conn, id)
			message, _ := bufio.NewReader(conn).ReadString('\n')
			fmt.Print("Client", id, " says: Received from server: "+message)
			outCh <- message
		}
	}()

	return
}

func receiveMessages(inCh chan string, conn net.Conn, id string) {
	for {
		select {
		case m := <-inCh:
			fmt.Println("Client inCh Received :", m)
			fmt.Fprintf(conn, id+" - "+m+"\n")
		default:
		}
	}
}
