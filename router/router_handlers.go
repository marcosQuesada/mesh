package router

import (
	"log"

	"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/peer"
	"github.com/marcosQuesada/mesh/router/handler"
)

func (r *defaultRouter) Handlers() map[message.MsgType]handler.Handler {
	return map[message.MsgType]handler.Handler{
		message.HELLO:   r.HandleHello,
		message.WELCOME: r.HandleWelcome,
		message.ACK:     r.HandleAck,
		message.ABORT:   r.HandleAbort,
		message.ERROR:   r.HandleError,
	}
}
func (r *defaultRouter) Notifiers() map[message.MsgType]bool{
	return map[message.MsgType]bool{
		message.HELLO:   false,
		message.WELCOME: true,
		message.ACK:     true,
		message.ABORT:   true,
		message.ERROR:   false,
	}
}

func (r *defaultRouter) Transactions() map[message.MsgType]bool{
	return map[message.MsgType]bool{
		message.HELLO:   true,
		message.WELCOME: true,
		message.ACK:     false,
		message.ABORT:   false,
		message.ERROR:   false,
	}
}

//HandleHello Request
func (r *defaultRouter) HandleHello(c peer.NodePeer, msg message.Message) (message.Message, error) {
	c.Identify(msg.(*message.Hello).From)
	if r.peerExists(c) {
		return &message.Abort{Id: msg.(*message.Hello).Id, From: r.from}, nil
	}

	c.State(peer.PeerStatusConnecting)

	return &message.Welcome{Id: msg.(*message.Hello).Id, From: r.from}, nil
}

//HandleWelcome Request
func (r *defaultRouter) HandleWelcome(c peer.NodePeer, msg message.Message) (message.Message, error) {
	err := r.accept(c)
	if err != nil {
		r.eventChan <- &peer.OnPeerErroredEvent{c.Node(), peer.PeerStatusError, err}

		return &message.Error{Id: msg.(*message.Welcome).Id, From: r.from}, err
	}
	c.State(peer.PeerStatusConnected)
	r.eventChan <- &peer.OnPeerConnectedEvent{c.Node(), peer.PeerStatusConnected, c.Mode()}

	go r.watcher.Watch(c)

	return &message.Ack{Id: msg.(*message.Welcome).Id, From: r.from}, nil

}

func (r *defaultRouter) HandleAbort(c peer.NodePeer, msg message.Message) (message.Message, error) {
	r.eventChan <- &peer.OnPeerAbortedEvent{c.Node(), peer.PeerStatusAbort}
	c.State(peer.PeerStatusAbort)
	c.Exit()

	return nil, nil
}

func (r *defaultRouter) HandleAck(c peer.NodePeer, msg message.Message) (message.Message, error) {
	err := r.accept(c)
	if err != nil {
		log.Println("Unexpected Error accepting Peer on HandleAck ", c.Node(), "msg:", msg)
	}

	c.State(peer.PeerStatusConnected)
	r.eventChan <- &peer.OnPeerConnectedEvent{msg.(*message.Ack).From, peer.PeerStatusConnected, c.Mode()}

	go r.watcher.Watch(c)

	return nil, nil
}

func (r *defaultRouter) HandleError(c peer.NodePeer, msg message.Message) (message.Message, error) {
	log.Println("HandleError ", c.Node(), "msg:", msg)

	return nil, nil
}
