package server

import (
	"fmt"
	"github.com/kuhufu/cm/protocol/Interface"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Channel struct {
	net.Conn
	srv           *Server
	id            string
	roomId        string
	status        int32
	outMsgQueue   chan Interface.Message
	outBytesQueue chan []byte //广播使用，避免消息多次encode
	exitC         chan struct{}
	closeOnce     sync.Once //保证连接只关闭一次
	CreateTime    time.Time //创建时间
	Metadata      sync.Map  //拓展信息可自由添加
	Network       string    //用什么协议连接的
	OnClose       func()    //close事件
}

func (c *Channel) Init(roomId string, channelId string) {
	c.roomId = roomId
	c.id = channelId
}

func NewChannel(conn net.Conn, network string, srv *Server) *Channel {
	c := &Channel{
		Conn:          conn,
		exitC:         make(chan struct{}),
		outMsgQueue:   make(chan Interface.Message, 4),
		outBytesQueue: make(chan []byte, 4),
		CreateTime:    time.Now(),
		Network:       network,
		srv:           srv,
	}

	return c
}

func (c *Channel) Close() error {
	var err error
	c.closeOnce.Do(func() {
		close(c.exitC)

		if c.OnClose != nil {
			c.OnClose()
		}

		err = c.Conn.Close()
		c.Empty()
	})
	return err
}

func (c *Channel) Exit() <-chan struct{} {
	return c.exitC
}

func (c *Channel) EnterOutMsg(msg Interface.Message) {
	select {
	case <-c.exitC:
		return
	case c.outMsgQueue <- msg:
	}
}

func (c *Channel) EnterOutBytes(data []byte) {
	select {
	case <-c.exitC:
		return
	case c.outBytesQueue <- data:
	}
}

func (c *Channel) WaitOutMsg() <-chan Interface.Message {
	return c.outMsgQueue
}

func (c *Channel) WaitOutBytes() <-chan []byte {
	return c.outBytesQueue
}

//清空消息，避免有goroutine阻塞在 outMsgQueue 或 outBytesQueue
func (c *Channel) Empty() {
	for {
		select {
		case msg := <-c.outMsgQueue:
			c.srv.GetMsgFactory().FreePoolMsg(msg)
		case <-c.outBytesQueue:

		default:
			if len(c.outMsgQueue) == 0 && len(c.outBytesQueue) == 0 {
				return
			}
		}
	}
}

func (c *Channel) StatusOk() bool {
	return atomic.LoadInt32(&c.status) == 0
}

func (c *Channel) String() string {
	return fmt.Sprintf("room: %v, channel_id: %v", c.roomId, c.id)
}

func (c *Channel) Id() string {
	return c.id
}

func (c *Channel) RoomId() string {
	return c.roomId
}
