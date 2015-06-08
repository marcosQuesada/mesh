package server

import (
	"github.com/marcosQuesada/mesh/cli"
	"github.com/marcosQuesada/mesh/cluster"
	"github.com/marcosQuesada/mesh/config"
	n "github.com/marcosQuesada/mesh/node"
	"github.com/marcosQuesada/mesh/peer"
	"log"
	"net"
)

type Server struct {
	config      *config.Config
	node        n.Node
	peerHandler peer.PeerHandler
	exit        chan bool
}

func New(c *config.Config) *Server {
	return &Server{
		config:      c,
		exit:        make(chan bool),
		node:        c.Addr,
		peerHandler: peer.DefaultPeerHandler(c.Addr),
	}
}

func (s *Server) Run() {
	defer close(s.exit)

	//@TODO: StartCoordinator
	o := cluster.StartOrchestrator(s.node, s.config.Cluster, s.peerHandler)
	go o.Run()

	s.startServer(o)

	for {
		select {
		case m := <-o.MainChan:
			log.Println("SERVER: Received Message on Main Channel ", m)
		case <-s.exit:
			//Notify Exit to remote Peer
			return
		}
	}
}

func (s *Server) Close() {
	s.exit <- true
}

func (s *Server) startServer(o *cluster.Orchestrator) {
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
			log.Println("Result from Accept is ", r)
		}
	}()
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
