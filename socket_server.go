package mesh

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

type Server struct {
	Port  int
	peers []*Peer
}

type Peer struct {
	Id   int
	Conn net.Conn
}

var clientId int

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

			peer := &Peer{Id: clientId, Conn: conn}
			s.Add(peer)
		}

	}()

	return
}

func (s *Server) Add(peer *Peer) {
	s.peers = append(s.peers, peer)
	go peer.handleSession()
	clientId++
}

func (s *Server) Broadcast() {
	go func() {
		time.Sleep(1 * time.Second)
		for _, p := range s.peers {
			fmt.Println("\nWriting Broadcast ", p.Id)
			p.Conn.Write([]byte("Broadcast from " + fmt.Sprintf("%d\n", p.Id)))
		}
	}()
}

func (p *Peer) handleSession() {
	for {
		// will listen for message to process ending in newline (\n)
		message, err := bufio.NewReader(p.Conn).ReadString('\n')
		if err != nil {
			fmt.Print("Error Receiving:", err)
			break
		}
		// output message received
		fmt.Print("Server Message Received:", string(message))
		p.Conn.Write([]byte(strings.ToUpper(message) + "\n"))
	}
}
