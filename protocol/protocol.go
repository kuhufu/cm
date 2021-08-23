package protocol

import (
	"github.com/kuhufu/cm/protocol/Interface"
	"github.com/kuhufu/cm/protocol/binary"
	"github.com/kuhufu/cm/protocol/json"
)

type MsgProto int

const (
	NONE = MsgProto(iota)
	BINARY
	JSON
	PROTOBUF
)

type MsgProtoFactory struct {
	NewMessage        func() Interface.Message
	NewDefaultMessage func() Interface.Message
	GetPoolMsg        func() Interface.Message
	FreePoolMsg       func(msg Interface.Message)
}

func GetFactory(msgProto MsgProto) *MsgProtoFactory {
	var factory *MsgProtoFactory
	switch msgProto {
	case BINARY:
		factory = &MsgProtoFactory{
			NewMessage:        binary.NewMessage,
			NewDefaultMessage: binary.NewDefaultMessage,
			GetPoolMsg:        binary.GetPoolMsg,
			FreePoolMsg:       binary.FreePoolMsg,
		}
	case JSON:
		factory = &MsgProtoFactory{
			NewMessage:        json.NewMessage,
			NewDefaultMessage: json.NewDefaultMessage,
			GetPoolMsg:        json.GetPoolMsg,
			FreePoolMsg:       json.FreePoolMsg,
		}
	case PROTOBUF:
		panic("unsupported")
	}

	return factory
}
