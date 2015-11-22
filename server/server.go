package server

import (
	"log"
	"net"

	//"github.com/marcosQuesada/mesh/cli"
	"github.com/marcosQuesada/mesh/cluster"
	"github.com/marcosQuesada/mesh/config"
	"github.com/marcosQuesada/mesh/dispatcher"
	n "github.com/marcosQuesada/mesh/node"
	"github.com/marcosQuesada/mesh/peer"
	"github.com/marcosQuesada/mesh/router"
"github.com/marcosQuesada/mesh/cli"
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
	}
}

// Start main server and all required dependencies
func (s *Server) Start() {
	disp := dispatcher.New()
	s.router = router.New(s.node, disp)
	c := cluster.Start(s.node, s.config.Cluster, disp)

	// Dispatcher definitions
	disp.RegisterListener(&peer.OnPeerConnectedEvent{}, c.OnPeerConnectedEvent)
	disp.RegisterListener(&peer.OnPeerDisconnectedEvent{}, c.OnPeerDisconnected)
	disp.RegisterListener(&peer.OnPeerDisconnectedEvent{}, s.OnPeerDisconnected)
	disp.RegisterListener(&peer.OnPeerAbortedEvent{}, c.OnPeerAborted)
	disp.RegisterListener(&peer.OnPeerErroredEvent{}, c.OnPeerErrored)

	// Start dispatcher channel consumers
	go disp.ConsumeEventChan()

	// Register Message Handlers
	s.router.RegisterHandlersFromInstance(c)
	s.router.RegisterHandlersFromInstance(c.Manager())

	// Aggregate coordinator send chan
	go s.router.AggregateChan(c.SndChan())

	//Boot Peer Server & Dial Peers
	s.startDialPeers()
	s.startPeerServer()

	//Boot Cli Server
	cli := cli.New(s.config.CliPort)
	cli.RegisterCommands(c)
	cli.RegisterCommands(s.router)
	go cli.Run()

	s.run()
}

//@TODO: Solve system shutdown
func (s *Server) Close() {
	s.router.Exit()
	close(s.exit)
}

// OnPeerDisconnected start peer recovering
func (s *Server) OnPeerDisconnected(e dispatcher.Event) {
	event := e.(*peer.OnPeerDisconnectedEvent)
	go s.initDialClient(event.Node)
}

//Used to shutdown server gracefully
func (s *Server) run() {
	for {
		select {
		case <-s.exit:
			//Notify Exit to remote Peer
			return
		}
	}
}

func (s *Server) startPeerServer() {
	listener, err := net.Listen("tcp", string(s.config.Addr.String()))
	if err != nil {
		log.Println("Error starting Socket Server: ", err)
		return
	}
	go s.startPeerAcceptor(listener)
}

func (s *Server) startPeerAcceptor(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error starting socket client to: ", s.node.String(), "err: ", err)
			return
		}

		p := peer.NewAcceptor(conn, s.node)
		go p.Run()

		s.router.Accept(p)
	}
}

//Start Dial Peers to all neighbours
func (s *Server) startDialPeers() {
	for _, node := range s.config.Cluster {
		//avoid local connexion
		if node.String() == s.node.String() {
			continue
		}
		go s.initDialClient(node)
	}
}

// InitDialClient starts a dial peer to a remote destination
func (s *Server)initDialClient(destination n.Node) {
	p, requestID := peer.InitDialClient(s.node, destination)

	// Register Hello request, wait and handle response
	go s.router.RequestListener().Register(requestID)
	s.router.Accept(p)
}
