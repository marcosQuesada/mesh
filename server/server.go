package server

import (
	//"io"
	"log"
	"net"
)

type Server struct {
	config        *Config
	clientHandler ClientHandler
	node          Node
	exit          chan bool
}

func New(c *Config) *Server {
	clientHandler := DefaultClientHandler()

	return &Server{
		clientHandler: clientHandler,
		config:        c,
		exit:          make(chan bool),
		node:          c.addr,
	}
}

func (s *Server) Run() {
	defer close(s.exit)

	// StartOrchestrator
	o := StartOrchestrator(s.node, s.config.cluster, s.clientHandler)
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

func (s *Server) startServer(o *Orchestrator) {
	listener, err := net.Listen("tcp", string(s.config.addr.String()))
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

			c := StartAcceptClient(conn, s.node)
			c.Run()

			r := o.Accept(c)
			log.Println("Result from Accept is ", r)
		}
	}()
}

// Cli Socket server
func (s *Server) startCliServer() error {
	listener, err := net.Listen("tcp", s.config.addr.String())
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
