package protocol

import (
	"github.com/kuhufu/cm/protocol/Interface"
	p "github.com/kuhufu/cm/protocol/binary"
)

func NewMessage() Interface.Message {
	return p.NewMessage()
}

func NewDefaultMessage() Interface.Message {
	return p.NewDefaultMessage()
}

func GetPoolMsg() Interface.Message {
	return p.GetPoolMsg()
}

func FreePoolMsg(msg Interface.Message) {
	p.FreePoolMsg(msg)
}
