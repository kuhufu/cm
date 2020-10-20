package server

import (
	"github.com/kuhufu/cm/protocol"
	"github.com/kuhufu/cm/protocol/Interface"
)

func CreateReplyMessage(srcMsg Interface.Message, data []byte) Interface.Message {
	msg := protocol.GetPoolMsg()
	msg.SetBody(data).SetCmd(srcMsg.Cmd()).SetRequestId(srcMsg.RequestId())
	return msg
}

type AuthReply struct {
	Ok        bool
	RoomId    string
	ChannelId string //不能为空，否则panic
	Data      []byte
	Metadata  map[interface{}]interface{}
	err       error
}
