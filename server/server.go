package server

import (
	//"io"
	"log"
	"net"
)

type Server struct {
	Router

	config        *Config
	clientHandler ClientHandler
	node          Node
	exit          chan bool
}

func New(c *Config) *Server {
	clientHandler := DefaultClientHandler()

	return &Server{
		Router:        NewRouter(c.addr, clientHandler),
		clientHandler: clientHandler,
		config:        c,
		exit:          make(chan bool),
		node:          c.addr,
	}
}

func (s *Server) Run() {
	defer close(s.exit)

	s.startServer()

	// StartOrchestrator
	st := StartOrchestrator(s.node, s.config.cluster, s.clientHandler)
	go st.Run()

	for {
		select {
		case <-s.exit:
			//Notify Exit to remote Peer
			//Shutdown peer connections
			return
		default:
		}
	}
}

func (s *Server) Close() {
	s.exit <- true
}

func (s *Server) startServer() {
	log.Print("Starting server: ", s.config.addr.String())

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
			go s.Router.Accept(c)
			go c.Run()
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
