package server

import (
	"fmt"
	"io"
	"log"
	"net"
)

type Server struct {
	Router

	config *Config

	node *Node
	exit chan bool
}

func New(c *Config) *Server {
	return &Server{
		Router: NewRouter(),

		config: c,
		exit:   make(chan bool),
		node:   c.raft_addr,
	}
}

func (s *Server) Run() {
	defer close(s.exit)

	s.startServer()

	for {
		select {
		case <-s.exit:
			//Notify Exit to remote Peer
			//Shutdown peer connections
			return
		default:
		}
	}
}

func (s *Server) Close() {
	s.exit <- true
}

func (s *Server) startServer() {
	log.Print("Starting server: ", s.config.raft_addr.Address())

	listener, err := net.Listen("tcp", string(s.config.raft_addr.Address()))
	if err != nil {
		log.Println("Error starting Socket Server: ", err)
		return
	}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Println("Error starting socket client to: ", s.node.Address(), "err: ", err)
				return
			}

			peer := NewJSONSocketPeer(conn)
			go s.handleConnection(peer)
			go s.Router.Accept(peer)

		}
	}()
}

func (s *Server) handleConnection(peer *SocketPeer) {
	defer peer.Conn.Close()
	defer close(peer.RcvChan)
	defer close(peer.exitChan)

	log.Print("Handling Connection from: ", peer.Id())

	for {
		msg, err := peer.Receive()
		if err != nil {
			if err != io.EOF {
				log.Print("Error reading connection ", err)
			}
			break
			//s.PeerHandler.Notify(peer.Id(), err)
		}

		fmt.Println("Received Message ", msg)
		peer.RcvChan <- msg
	}
}

// Cli Socket server
func (s *Server) startCliServer() error {
	listener, err := net.Listen("tcp", string(s.config.addr.Address()))
	if err != nil {
		log.Println("Error Listening Cli Server")
		return err
	}
	go func() error {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Println("Error Accepting")
				return err
			}
			defer conn.Close()
			go s.handleCliConnection(conn)
		}
	}()

	return nil
}

// Socket Client access
func (s *Server) handleCliConnection(c net.Conn) {
	defer c.Close()

	cli := &CliSession{
		conn:   c,
		server: s,
	}
	cli.handle()
}
