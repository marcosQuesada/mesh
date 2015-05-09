package server

import (
	"log"
	"net"
)

type Server struct {
	config   *Config
	exit     chan bool
	peers    []*Peer         //clients that handle remotes
	links    map[string]Link //connections served
	listener net.Listener
	raft     net.Listener
	node     *Node
}

func New(c *Config) *Server {
	return &Server{
		config: c,
		exit:   make(chan bool),
		node:   c.raft_addr,
		peers:  make([]*Peer, 0),
		links:  make(map[string]Link),
	}
}

func (s *Server) Run() {
	defer log.Println("Server Exitting")
	defer close(s.exit)

	s.startServer()
	s.startPeer()
	for {
		select {
		case <-s.exit:
			//Notify Exit to remote Peer
			//Shutdown peer connections
			s.exitPeer()
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

	s.listener, err = net.Listen("tcp", s.config.addr.Address())
	go func() error {
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				log.Println("Error Accepting")
				return err
			}
			defer conn.Close()
			go s.handleCliSession(conn)
		}
	}()

	s.raft, err = net.Listen("tcp", s.config.raft_addr.Address())
	go func() error {
		for {
			srvLink := NewServerLink(s.raft)
			srvLink.Connect(&Node{})
			defer srvLink.Exit()

			go s.handleClusterConnection(srvLink)

		}
	}()

	return err
}

func (s *Server) startPeer() {
	for _, p := range s.config.raft_cluster {

		//if destination is not local
		if p.Address() != s.node.Address() {
			Peer := NewPeer(s.node, p, 1000)
			Peer.Run()

			s.peers = append(s.peers, Peer)
		}
	}
}

func (s *Server) exitPeer() {
	for _, peer := range s.peers {
		peer.Exit()
	}
}

//Intra cluster socket server
func (s *Server) handleClusterConnection(c Link) {
	var first bool = true // first message belongs to remote id
	log.Println("Handling Raft connection ")

	for {
		byteMessage, err := c.Receive()
		if err != nil {
			log.Println("Error receiving message")

			return
		}
		message := string(byteMessage)

		// Identify from remote node
		if first {
			remoteNode, err := parse(message)
			if err != nil {
				log.Println("First message is NOT ID")
				continue
			}
			first = false
			c.Identify(remoteNode)
			s.addLink(c)
			defer s.removeLink(remoteNode)
		}

		log.Println("Server received from remote node: ", c.Id().Address(), "message: ", message)
	}
}

func (s *Server) addLink(p Link) {
	s.links[p.Id().Address()] = p
}

func (s *Server) removeLink(remoteId *Node) {
	delete(s.links, remoteId.Address())
}

func (s *Server) Links() map[string]Link {
	return s.links
}

// Socket Client access
func (s *Server) handleCliSession(c net.Conn) {
	defer c.Close()

	cli := &CliSession{
		conn:   c,
		server: s,
	}
	cli.handle()
}
