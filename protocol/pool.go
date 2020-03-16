package protocol

import "sync"

var pool = sync.Pool{
	New: func() interface{} {
		return NewMessageWithDefault()
	},
}

func GetPoolMsg() *Message {
	msg := pool.Get().(*Message)
	return msg
}

func FreePoolMsg(msg *Message) {
	pool.Put(msg)
}
