package binary

import "sync"

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
