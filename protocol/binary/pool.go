package binary

import (
	"github.com/kuhufu/cm/protocol/Interface"
	"sync"
)

var pool = sync.Pool{
	New: func() interface{} {
		return NewDefaultMessage()
	},
}

func GetPoolMsg() Interface.Message {
	msg := pool.Get().(*Message)
	return msg
}

func FreePoolMsg(msg Interface.Message) {
	pool.Put(msg)
}
