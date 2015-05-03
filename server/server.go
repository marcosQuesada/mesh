package server

import (
	"bufio"
	"fmt"
	"net"
)

type Server struct {
	config   *Config
	exit     chan bool
	peers    []*Peer
	listener net.Listener
	raft     net.Listener
	node     *Node
}

func New(c *Config) *Server {
	return &Server{
		config: c,
		exit:   make(chan bool),
		node:   c.raft_addr,
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
	s.listener, err = net.Listen("tcp", s.config.addr.Address())
	go func() error {
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				fmt.Println("Error Accepting")
				return err
			}
			defer conn.Close()
			go s.handleConnection(conn)
		}
	}()

	s.raft, err = net.Listen("tcp", s.config.raft_addr.Address())
	go func() error {
		for {
			conn, err := s.raft.Accept()
			if err != nil {
				fmt.Println("Error Accepting")
				return err
			}
			defer conn.Close()
			go s.handleRaftConnection(conn)
		}
	}()

	return err
}

func (s *Server) startPeers() {
	fmt.Println("Local peer ", s.node.Address())
	for _, p := range s.config.raft_cluster {
		if p.Address() != s.node.Address() {
			fmt.Println("Creating peer ", p.Address())
			peer := NewPeer(s.node, p, 1000)
			peer.Run()
			s.peers = append(s.peers, peer)
		}
	}
}

func (s *Server) exitPeers() {
	for _, peer := range s.peers {
		peer.Exit()
	}
}

func (s *Server) handleConnection(c net.Conn) {
	fmt.Println("Handling connection ")
	defer c.Close()
	for {
		message, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Print("Error Receiving on server, err ", err)
			return
		}

		fmt.Println("Server received Message ", message)
	}
}

func (s *Server) handleRaftConnection(c net.Conn) {
	fmt.Println("Handling Raft connection ")
	defer c.Close()
	for {
		message, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Print("Error Receiving on server, err ", err)
			return
		}

		fmt.Println("Server received Message ", message)
	}

}
