package json

import (
	"github.com/kuhufu/cm/protocol/consts"
	"github.com/kuhufu/cm/protocol/global"
	"sync"
)

func init() {
	global.Protocol = consts.JSON
}

var pool = sync.Pool{
	New: func() interface{} {
		return NewMessage()
	},
}

func GetPoolMsg() *MessageV1 {
	msg := pool.Get().(*MessageV1)
	return msg
}

func FreePoolMsg(msg interface{}) {
	pool.Put(msg)
}
