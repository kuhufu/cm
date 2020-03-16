package server

import (
	"github.com/kuhufu/cm/protocol"
	"time"
)

func CreateReplyMessage(srcMsg *protocol.Message, data []byte) *protocol.Message {
	msg := protocol.GetPoolMsg()
	msg.SetBody(data).SetCmd(srcMsg.Cmd()).SetRequestId(srcMsg.RequestId())
	return msg
}

type AuthReply struct {
	Ok       bool
	ConnId   string //不能为空，否则panic
	GroupIds []string
	Data     []byte
	AuthTime time.Time
	Extends  map[interface{}]interface{}
	err      error
}
