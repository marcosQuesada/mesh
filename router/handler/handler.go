package handler

import(
	"github.com/marcosQuesada/mesh/peer"
	"github.com/marcosQuesada/mesh/message"
)

//Handler represent a method to be invoked with message
//error will be returned on unexpected handling
type Handler func(peer.NodePeer, message.Message) (message.Message, error)

type MessageHandler interface {
	Handlers() map[message.MsgType]Handler
}