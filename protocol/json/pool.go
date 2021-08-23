package json

import (
	"github.com/kuhufu/cm/protocol/Interface"
	"sync"
)

var pool = sync.Pool{
	New: func() interface{} {
		return NewMessage()
	},
}

func GetPoolMsg() Interface.Message {
	msg := pool.Get().(*MessageV1)
	return msg
}

func FreePoolMsg(msg Interface.Message) {
	if _, ok := msg.(*MessageV1); !ok {
		panic("ddddddddddd")
	}
	pool.Put(msg)
}
