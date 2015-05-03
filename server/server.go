package server

import (
	"bufio"
	"fmt"
	"net"
)

type Server struct {
	config     *Config
	exit       chan bool
	peerClient []*PeerClient    //clients that handle remotes
	peers      map[string]*Node //Connected from server
	listener   net.Listener
	raft       net.Listener
	node       *Node
}

func New(c *Config) *Server {
	return &Server{
		config: c,
		exit:   make(chan bool),
		node:   c.raft_addr,
		peers:  make(map[string]*Node),
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

func (s *Server) handleConnection(c net.Conn) {
	fmt.Println("Handling connection ")
	defer c.Close()
	var first bool = true // forst message belongs to remote id

	for {
		fmt.Println("Before Server first ", first)
		message, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Print("Error Receiving on server, err ", err)
			return
		}
		fmt.Println("Server first ", first)
		if first {
			fmt.Println("First message", message)
			remoteNode, err := parse(message)
			if err != nil {
				fmt.Println("First message is NOT ID")
				continue
			}
			fmt.Println("Remote node id : ", remoteNode)
		}

		fmt.Println("Server received Message ", message)
		first = false
	}
}

func (s *Server) handleRaftConnection(c net.Conn) {
	var first bool = true // forst message belongs to remote id
	var remoteId *Node
	fmt.Println("Handling Raft connection ")
	defer c.Close()

	for {
		message, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Print("Error Receiving on server, err ", err)
			return
		}
		messageParts := clear(message)
		if len(messageParts) != 0 {
			message = messageParts[0]
		}
		if first {
			//Clear end character \n
			remoteNode, err := parse(message)
			if err != nil {
				fmt.Println("First message is NOT ID")
				continue
			}
			//Registering Peer
			remoteId = remoteNode
			s.addPeer(remoteId)
			defer s.removePeer(remoteId)
		}

		fmt.Println("Server received Message ", message, "from remote node: ", remoteId.Address())
		first = false
	}
}

func (s *Server) addPeer(remoteId *Node) {
	s.peers[remoteId.Address()] = remoteId
	fmt.Println("Total Peers are ", len(s.peers))
}

func (s *Server) removePeer(remoteId *Node) {
	delete(s.peers, remoteId.Address())
	fmt.Println("Total Peers are ", len(s.peers))
}
