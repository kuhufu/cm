package server

import "sync/atomic"

type Channel struct {
	*Conn
	Id         ChannelId
	ClientType ClientType
	RoomId     RoomId
	status     int32
}

func (c *Channel) StatusOk() bool {
	return atomic.LoadInt32(&c.status) == 0
}
