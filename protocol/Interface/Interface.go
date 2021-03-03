package Interface

import (
	"github.com/kuhufu/cm/protocol/consts"
	"io"
)

type Message interface {
	io.ReaderFrom
	io.WriterTo

	Encode() []byte

	Cmd() Cmd
	Body() []byte
	RequestId() uint32

	SetCmd(Cmd) Message
	SetBody([]byte) Message
	SetRequestId(uint32) Message
}

type Cmd uint32

var cmdMap = map[Cmd]string{
	consts.CmdUnknown:    "CmdUnknown",
	consts.CmdAuth:       "CmdAuth",
	consts.CmdPush:       "CmdPush",
	consts.CmdHeartbeat:  "CmdHeartbeat",
	consts.CmdClose:      "CmdClose",
	consts.CmdServerPush: "CmdServerPush",
}

func (c Cmd) String() string {
	return cmdMap[c]
}
