package server

import (
	"github.com/kuhufu/cm/protocol"
	"github.com/kuhufu/cm/protocol/Interface"
	"net"
	"sync"
	"time"
)

type Conn struct {
	net.Conn
	outMsgQueue   chan Interface.Message
	outBytesQueue chan []byte //广播使用，避免消息多次encode
	exitC         chan struct{}

	closeOnce  sync.Once //保证连接只关闭一次
	CreateTime time.Time //创建时间
	Metadata   sync.Map  //拓展信息可自由添加
	OnClose    func()    //close事件
}

func NewConn(conn net.Conn) *Conn {
	c := &Conn{
		Conn:          conn,
		exitC:         make(chan struct{}),
		outMsgQueue:   make(chan Interface.Message, 4),
		outBytesQueue: make(chan []byte, 4),
		CreateTime:    time.Now(),
	}

	//创建时间+内存地址，本地测试中创建时间可能会相同。在创建时间相同的情况下，内存地址必不相同
	return c
}

func (conn *Conn) Close() error {
	var err error
	conn.closeOnce.Do(func() {
		close(conn.exitC)

		if conn.OnClose != nil {
			conn.OnClose()
		}

		err = conn.Conn.Close()
		conn.Empty()
	})
	return err
}

func (conn *Conn) Exit() <-chan struct{} {
	return conn.exitC
}

//消息是否需要完整写入
func (conn *Conn) MessageNeedFullWrite() bool {
	if v, ok := conn.Conn.(Interface.NeedFullWrite); ok {
		return v.MessageNeedFullWrite()
	}
	return false
}

func (conn *Conn) EnterOutMsg(msg Interface.Message) {
	select {
	case <-conn.exitC:
		return
	case conn.outMsgQueue <- msg:
	}
}

func (conn *Conn) EnterOutBytes(data []byte) {
	select {
	case <-conn.exitC:
		return
	case conn.outBytesQueue <- data:
	}
}

func (conn *Conn) WaitOutMsg() <-chan Interface.Message {
	return conn.outMsgQueue
}

func (conn *Conn) WaitOutBytes() <-chan []byte {
	return conn.outBytesQueue
}

//清空消息，避免有goroutine阻塞在 outMsgQueue 或 outBytesQueue
func (conn *Conn) Empty() {
	for {
		select {
		case msg := <-conn.outMsgQueue:
			protocol.FreePoolMsg(msg)
		case <-conn.outBytesQueue:

		default:
			if len(conn.outMsgQueue) == 0 && len(conn.outBytesQueue) == 0 {
				return
			}
		}
	}
}
