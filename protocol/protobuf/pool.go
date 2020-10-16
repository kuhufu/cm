package protobuf

import "sync"

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
