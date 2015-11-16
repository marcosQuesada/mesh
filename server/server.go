package server

import (
	"log"
	"net"

	"github.com/marcosQuesada/mesh/cli"
	"github.com/marcosQuesada/mesh/cluster"
	"github.com/marcosQuesada/mesh/config"
	"github.com/marcosQuesada/mesh/dispatcher"
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
	c := cluster.Start(s.node, s.config.Cluster)

	disp := dispatcher.New()
	disp.RegisterListener(&peer.OnPeerConnectedEvent{}, c.OnPeerConnectedEvent)
	disp.RegisterListener(&peer.OnPeerDisconnectedEvent{}, c.OnPeerDisconnected)
	disp.RegisterListener(&peer.OnPeerDisconnectedEvent{}, s.OnPeerDisconnected)
	disp.RegisterListener(&peer.OnPeerAbortedEvent{}, c.OnPeerAborted)
	disp.RegisterListener(&peer.OnPeerErroredEvent{}, c.OnPeerErrored)
	go disp.ConsumeEventChan()
	go disp.AggregateChan(s.router.Events())

	s.router.RegisterHandlersFromInstance(c)
	s.router.RegisterHandlersFromInstance(c.Manager())
	//aggregate coordinator snd chan
	go s.router.AggregateChan(c.SndChan())

	s.startDialPeers()
	s.startServer()
	s.run()

	//Boot Cli Server
	cli := cli.New(s.config.Addr.Port)
	cli.RegisterCommands(c)
	cli.RegisterCommands(s.router)
	go cli.Run()

	//TODO: aggregate watcher snd chan
	//s.router.AggregateChan(c.SndChan())

}

func (s *Server) Close() {
	//TODO: fix dispatcher Shutdown
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

		go s.InitDialClient(node)
	}
}

func (s *Server) InitDialClient(destination n.Node) {
	p, requestID := peer.InitDialClient(s.node, destination)
	go s.router.RequestListener().Register(requestID)
	s.router.Accept(p)
}

func (s *Server) OnPeerDisconnected (e dispatcher.Event) {
	event := e.(*peer.OnPeerDisconnectedEvent)
	go s.InitDialClient(event.Node)
}