package watch

import (
	"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/peer"
	"github.com/marcosQuesada/mesh/router/handler"
)

func (w *defaultWatcher) Handlers() map[message.MsgType]handler.Handler {
	return map[message.MsgType]handler.Handler{
		message.PING: w.HandlePing,
		message.PONG: w.HandlePong,
	}
}

func (w *defaultWatcher) Notifiers() map[message.MsgType]bool {
	return map[message.MsgType]bool{
		message.PING: false,
		message.PONG: true,
	}
}

func (w *defaultWatcher) Transactions() map[message.MsgType]bool {
	return map[message.MsgType]bool{
		message.PING: true,
		message.PONG: false,
	}
}

func (w *defaultWatcher) HandlePing(c peer.NodePeer, msg message.Message) (message.Message, error) {
	ping := msg.(*message.Ping)

	return &message.Pong{Id: ping.Id, From: ping.To, To: ping.From}, nil
}

func (w *defaultWatcher) HandlePong(c peer.NodePeer, msg message.Message) (message.Message, error) {
	return nil, nil
}
