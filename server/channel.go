package server

import "sync/atomic"

type Channel struct {
	*Conn
	Id         ChannelId
	ClientType ClientType
	status     int32
}

func (c *Channel) StatusOk() bool {
	return atomic.LoadInt32(&c.status) == 0
}
