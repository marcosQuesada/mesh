package server

import (
	"fmt"
	"net"
)

type Server struct {
	config     *Config
	exit       chan bool
	peerClient []*PeerClient //clients that handle remotes
	peerList   map[string]Peer
	//peers      map[string]*Node //Connected from server
	listener net.Listener
	raft     net.Listener
	node     *Node
}

func New(c *Config) *Server {
	return &Server{
		config: c,
		exit:   make(chan bool),
		node:   c.raft_addr,
		//peers:    make(map[string]*Node),
		peerList: make(map[string]Peer),
	}
}

func (s *Server) Run() {
	defer fmt.Println("Server Exitting")
	defer close(s.exit)

	s.startServer()
	s.startPeerClient()
	for {
		select {
		case <-s.exit:
			//Notify Exit to remote PeerClient
			//Shutdown peer connections
			s.exitPeerClient()
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
				fmt.Println("Error Accepting")
				return err
			}
			defer conn.Close()
			go s.handleCliSession(conn)
		}
	}()

	s.raft, err = net.Listen("tcp", s.config.raft_addr.Address())
	go func() error {
		for {
			srvPeer := NewServerPeer(s.raft)
			srvPeer.Connect(&Node{})
			defer srvPeer.Exit()

			go s.handleClusterConnection(srvPeer)

		}
	}()

	return err
}

func (s *Server) startPeerClient() {
	for _, p := range s.config.raft_cluster {

		//if destination is not local
		if p.Address() != s.node.Address() {
			PeerClient := NewPeerClient(s.node, p, 1000)
			PeerClient.Run()

			s.peerClient = append(s.peerClient, PeerClient)
		}
	}
}

func (s *Server) exitPeerClient() {
	for _, peer := range s.peerClient {
		peer.Exit()
	}
}

//Intra cluster socket server
func (s *Server) handleClusterConnection(c Peer) {
	var first bool = true // first message belongs to remote id
	fmt.Println("Handling Raft connection ")

	for {
		byteMessage, err := c.Receive()
		if err != nil {
			return
		}
		message := string(byteMessage)

		// Identify from remote node
		if first {
			remoteNode, err := parse(message)
			if err != nil {
				fmt.Println("First message is NOT ID")
				continue
			}
			first = false
			c.Identify(remoteNode)
			s.addPeer(c)
			defer s.removePeer(remoteNode)
		}

		fmt.Println("Server received from remote node: ", c.Id().Address(), "message: ", message)
	}
}

func (s *Server) addPeer(p Peer) {
	s.peerList[p.Id().Address()] = p
}

func (s *Server) removePeer(remoteId *Node) {
	delete(s.peerList, remoteId.Address())
}

func (s *Server) PeerList() map[string]Peer {
	return s.peerList
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
