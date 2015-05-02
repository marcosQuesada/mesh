package mesh

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

type Server struct {
	Port   int
	peers  []*Peer
	finish bool
}

type Peer struct {
	Id     int
	Conn   net.Conn
	Finish bool
}

var clientId int

func (s *Server) StartServer() (done chan bool) {
	done = make(chan bool)
	fmt.Println("Launching server...")
	// listen on all interfaces
	go func() {
		done <- true
		ln, _ := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))

		for !s.finish {
			conn, err := ln.Accept()
			defer conn.Close()

			if err != nil {
				fmt.Println("Error Accepting")
			}

			peer := &Peer{Id: clientId, Conn: conn}
			s.Add(peer)
		}

		for _, p := range s.peers {
			p.Finish = true
		}

	}()

	return
}

func (s *Server) Add(peer *Peer) {
	s.peers = append(s.peers, peer)
	go peer.handleSession()
	clientId++
}

func (s *Server) Terminate() {
	s.finish = true
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
	defer p.Conn.Close()
	for !p.Finish {
		// will listen for message to process ending in newline (\n)
		message, err := bufio.NewReader(p.Conn).ReadString('\n')
		if err != nil {
			fmt.Print("Error Receiving on node:", fmt.Sprintf("%d\n", p.Id), "err ", err)
			return
			//break
		}
		// output message received
		fmt.Print("Server Message Received:", string(message))
		p.Conn.Write([]byte(strings.ToUpper(message) + "\n"))
	}
}
