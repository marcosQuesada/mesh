package server

import (
	"log"
	"net"

	"github.com/marcosQuesada/mesh/cli"
	"github.com/marcosQuesada/mesh/cluster"
	"github.com/marcosQuesada/mesh/config"
	"github.com/marcosQuesada/mesh/dispatcher"
	"github.com/marcosQuesada/mesh/message"
	n "github.com/marcosQuesada/mesh/node"
	"github.com/marcosQuesada/mesh/peer"
	"github.com/marcosQuesada/mesh/router"
)

type Server struct {
	config *config.Config
	node   n.Node
	exit   chan bool
	router router.Router
}

func New(c *config.Config) *Server {
	return &Server{
		config: c,
		exit:   make(chan bool),
		node:   c.Addr,
		router: router.New(c.Addr),
	}
}

func (s *Server) Start() {
	c := cluster.StartCoordinator(s.node, s.config.Cluster)
	go c.Run()

	s.router.RegisterHandler(message.COMMAND, c.HandleCommand)

	d := dispatcher.New()
	d.RegisterListener(&peer.OnPeerConnectedEvent{}, c.OnPeerConnectedEvent)
	d.RegisterListener(&peer.OnPeerDisconnectedEvent{}, c.OnPeerDisconnected)
	d.RegisterListener(&peer.OnPeerAbortedEvent{}, c.OnPeerAborted)
	d.RegisterListener(&peer.OnPeerErroredEvent{}, c.OnPeerErrored)

	go d.Run()
	go d.Aggregate(s.router.Events())

	s.startDialPeers()
	s.startServer()
	s.run()
}

func (s *Server) Close() {
	//d.Exit()

	close(s.exit)
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

	go s.startAcceptorPeers(listener)
}

func (s *Server) startAcceptorPeers(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error starting socket client to: ", s.node.String(), "err: ", err)
			return
		}

		c := peer.NewAcceptor(conn, s.node)
		go c.Run()

		s.router.Accept(c)
	}
}

//Start Dial Peers
func (s *Server) startDialPeers() {
	for _, node := range s.config.Cluster {
		//avoid local connexion
		if node.String() == s.node.String() {
			continue
		}

		go s.router.InitDialClient(node)
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
