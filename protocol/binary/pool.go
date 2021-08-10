package binary

import (
	"github.com/kuhufu/cm/protocol/consts"
	"github.com/kuhufu/cm/protocol/global"
	"sync"
)

func init() {
	global.Protocol = consts.BINARY
}

var pool = sync.Pool{
	New: func() interface{} {
		return NewDefaultMessage()
	},
}

func GetPoolMsg() *Message {
	msg := pool.Get().(*Message)
	return msg
}

func FreePoolMsg(msg interface{}) {
	pool.Put(msg)
}
