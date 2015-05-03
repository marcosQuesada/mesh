package server

import (
	"fmt"
	"net"
)

type Server struct {
	config   *Config
	exit     chan bool
	peers    []*Peer
	listener net.Listener
	raft     net.Listener
}

func New(c *Config) *Server {
	return &Server{
		config: c,
		exit:   make(chan bool),
	}
}

func (s *Server) Run() {
	defer fmt.Println("Server Exitting")
	defer close(s.exit)

	s.startServer()
	s.startPeers()
	for {
		select {
		case <-s.exit:
			//Notify Exit to remote peers
			//Shutdown peer connections
			s.exitPeers()
			return
		default:
		}
	}
}

func (s *Server) Close() {
	s.exit <- true
}

func (s *Server) startServer() error {
	var err error

	//@TODO! THink about that
	s.listener, err = net.Listen("tcp", s.config.addr)
	go func() error {
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				fmt.Println("Error Accepting")
				return err
			}
			defer conn.Close()
			go handleConnection(conn)
		}
	}()

	s.raft, err = net.Listen("tcp", s.config.raft_addr)
	go func() error {
		for {
			conn, err := s.raft.Accept()
			if err != nil {
				fmt.Println("Error Accepting")
				return err
			}
			defer conn.Close()
			go handleRaftConnection(conn)
		}
	}()

	return err
}

func (s *Server) startPeers() {
	for _, p := range s.config.raft_cluster {
		peer := NewPeer(p, 1000)
		peer.Run()
		s.peers = append(s.peers, peer)
	}
}

func (s *Server) exitPeers() {
	for _, peer := range s.peers {
		peer.Exit()
	}
}

func handleConnection(c net.Conn) {

}

func handleRaftConnection(c net.Conn) {

}
