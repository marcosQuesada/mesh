package server

import (
	"fmt"
)

type Server struct {
	config *Config
	exit   chan bool
	peers  []*Peer
}

func New(c *Config) *Server {
	fmt.Println("Config is ", c)
	return &Server{
		config: c,
		exit:   make(chan bool),
	}
}

func (s *Server) Run() {
	defer fmt.Println("Server Exitting")
	defer close(s.exit)

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
