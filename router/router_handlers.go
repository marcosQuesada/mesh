package router

import (
	"log"

	"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/peer"
)

//HandleHello Request
func (r *defaultRouter) HandleHello(c peer.NodePeer, msg message.Message) (message.Message, error) {
	c.Identify(msg.(*message.Hello).From)
	err := r.accept(c)
	if err != nil {
		r.eventChan <- &peer.OnPeerAbortedEvent{msg.(*message.Hello).From, peer.PeerStatusError}

		return &message.Abort{Id: msg.(*message.Hello).Id, From: r.from}, nil
	}
	c.State(peer.PeerStatusConnecting)
	r.eventChan <- &peer.OnPeerConnectedEvent{msg.(*message.Hello).From, peer.PeerStatusConnected, c.Mode()}

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

	return nil, nil

}

func (r *defaultRouter) HandleAbort(c peer.NodePeer, msg message.Message) (message.Message, error) {
	r.eventChan <- &peer.OnPeerAbortedEvent{c.Node(), peer.PeerStatusAbort}
	c.State(peer.PeerStatusAbort)
	c.Exit()

	return nil, nil
}

func (r *defaultRouter) HandleAck(c peer.NodePeer, msg message.Message) (message.Message, error) {
	c.State(peer.PeerStatusConnected)

	return nil, nil
}

func (r *defaultRouter) HandleError(c peer.NodePeer, msg message.Message) (message.Message, error) {
	log.Println("HandleError ", c.Node(), "msg:", msg)

	return nil, nil
}
