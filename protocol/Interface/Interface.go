package Interface

import (
	"io"
)

type Message interface {
	io.ReaderFrom
	io.WriterTo

	Encode() []byte

	Cmd() uint32
	Body() []byte
	RequestId() uint32

	SetCmd(uint32) Message
	SetBody([]byte) Message
	SetRequestId(uint32) Message
}

//标记接口，需要一次性写入完整消息，否则头和body将分开写
type NeedFullWrite interface {
	MessageNeedFullWrite() bool
}
