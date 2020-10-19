package server

import (
	"fmt"
	"github.com/kuhufu/cm/protocol"
	"github.com/kuhufu/cm/protocol/Interface"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type ClientType = string
type ChannelId = string

type Channel struct {
	net.Conn
	Id            ChannelId
	ClientType    ClientType
	RoomId        RoomId
	status        int32
	outMsgQueue   chan Interface.Message
	outBytesQueue chan []byte //广播使用，避免消息多次encode
	exitC         chan struct{}
	closeOnce     sync.Once //保证连接只关闭一次
	CreateTime    time.Time //创建时间
	Metadata      sync.Map  //拓展信息可自由添加
	OnClose       func()    //close事件
}

func (c *Channel) Init(roomId RoomId, clientType ClientType) {
	c.RoomId = roomId
	c.ClientType = clientType
	c.Id = fmt.Sprintf("%v_%v", roomId, clientType)
}

func NewChannel(conn net.Conn) *Channel {
	c := &Channel{
		Conn:          conn,
		exitC:         make(chan struct{}),
		outMsgQueue:   make(chan Interface.Message, 4),
		outBytesQueue: make(chan []byte, 4),
		CreateTime:    time.Now(),
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

//消息是否需要完整写入
func (c *Channel) MessageNeedFullWrite() bool {
	if v, ok := c.Conn.(Interface.NeedFullWrite); ok {
		return v.MessageNeedFullWrite()
	}
	return false
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
			protocol.FreePoolMsg(msg)
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
	return fmt.Sprintf("room:%v, client_type:%v", c.RoomId, c.ClientType)
}
