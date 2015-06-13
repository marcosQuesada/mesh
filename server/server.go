package server

import (
	"github.com/marcosQuesada/mesh/cli"
	"github.com/marcosQuesada/mesh/cluster"
	"github.com/marcosQuesada/mesh/config"
	"github.com/marcosQuesada/mesh/dispatcher"
	n "github.com/marcosQuesada/mesh/node"
	"github.com/marcosQuesada/mesh/peer"
	"github.com/marcosQuesada/mesh/peer_handler"
	"log"
	"net"
)

type Server struct {
	config      *config.Config
	node        n.Node
	peerHandler peer_handler.PeerHandler
	exit        chan bool
}

func New(c *config.Config) *Server {
	return &Server{
		config:      c,
		exit:        make(chan bool),
		node:        c.Addr,
		peerHandler: peer_handler.DefaultPeerHandler(c.Addr),
	}
}

func (s *Server) Start() {
	defer close(s.exit)

	c := cluster.StartCoordinator(s.node, s.config.Cluster, s.peerHandler)
	c.Run()

	d := dispatcher.New()
	d.RegisterListener(&peer.OnPeerConnectedEvent{}, c.OnPeerConnectedEvent)
	d.RegisterListener(&peer.OnPeerDisconnectedEvent{}, c.OnPeerDisconnected)
	d.RegisterListener(&peer.OnPeerAbortedEvent{}, c.OnPeerAborted)
	d.RegisterListener(&peer.OnPeerErroredEvent{}, c.OnPeerErrored)

	d.Run()
	d.Aggregate(s.peerHandler.Events())

	s.startClientPeers()
	s.startServer()
	s.run()
}

func (s *Server) Close() {
	s.exit <- true
}

func (s *Server) run() {
	for {
		select {
		case <-s.exit:
			//Notify Exit to remote Peer
			return
		}
	}
}

func (s *Server) startServer() {
	listener, err := net.Listen("tcp", string(s.config.Addr.String()))
	if err != nil {
		log.Println("Error starting Socket Server: ", err)
		return
	}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Println("Error starting socket client to: ", s.node.String(), "err: ", err)
				return
			}

			c := peer.NewAcceptor(conn, s.node)
			c.Run()

			r := s.peerHandler.Handle(c)
			rn := c.Node()
			log.Println("Server link from:", rn.String(), " result: ", r)
		}
	}()
}

func (s *Server) startClientPeers() {
	//Start Dial Peers
	for _, node := range s.config.Cluster {
		//avoid local connexion
		if node.String() == s.node.String() {
			continue
		}

		go func(destination n.Node) {
			log.Println("Starting Dial Client from Node ", s.node.String(), "destination: ", node.String())
			//Blocking call, wait until connection success
			c := peer.NewDialer(s.node, destination)
			c.Run()
			//Say Hello and wait response
			c.SayHello()
			r := s.peerHandler.Handle(c)
			rn := c.Node()
			log.Println("Dial link to to:", rn.String(), "result: ", r)
		}(node)
	}
}

// Cli Socket server
func (s *Server) startCliServer() error {
	listener, err := net.Listen("tcp", s.config.Addr.String())
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
func (s *Server) handleCliConnection(conn net.Conn) {
	defer conn.Close()

	c := &cli.CliSession{
		Conn: conn,
		//server: s,
	}
	c.Handle()
}
