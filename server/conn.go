package server

import (
	"fmt"
	"net"
	"sync"
	"time"
	"unsafe"

	"github.com/kuhufu/cm/protocol"
)

type Conn struct {
	net.Conn
	Id           string
	userId       string
	outMsgChan   chan *protocol.Message
	outBytesChan chan []byte //广播使用，避免消息多次encode
	exitChan     chan struct{}

	closeOnce sync.Once //保证连接只关闭一次
	version   string    //创建连接时的版本号
	Metadata  sync.Map  //拓展信息可自由添加
}

func NewConn(conn net.Conn) *Conn {
	c := &Conn{
		Conn:         conn,
		exitChan:     make(chan struct{}),
		outMsgChan:   make(chan *protocol.Message, 4),
		outBytesChan: make(chan []byte, 4),
		closeOnce:    sync.Once{},
		Metadata:     sync.Map{},
	}

	//创建时间+内存地址，本地测试中创建时间可能会相同。在创建时间相同的情况下，内存地址必不相同
	c.version = fmt.Sprintf("%x-%x", time.Now().UnixNano(), uintptr(unsafe.Pointer(c)))
	return c
}

func (conn *Conn) Init(userId, connId string) {
	conn.userId = userId
	conn.Id = connId
}

func (conn *Conn) Close() error {
	var err error
	conn.closeOnce.Do(func() {
		close(conn.exitChan)
		err = conn.Conn.Close()
		conn.Empty()
	})
	return err
}

func (conn *Conn) Exit() <-chan struct{} {
	return conn.exitChan
}

//消息是否需要完整写入
func (conn *Conn) MessageNeedFullWrite() bool {
	if v, ok := conn.Conn.(protocol.NeedFullWrite); ok {
		return v.MessageNeedFullWrite()
	}
	return false
}

func (conn *Conn) EnterOutMsg(msg *protocol.Message) {
	select {
	case <-conn.exitChan:
		return
	case conn.outMsgChan <- msg:
	}
}

func (conn *Conn) EnterOutBytes(data []byte) {
	select {
	case <-conn.exitChan:
		return
	case conn.outBytesChan <- data:
	}
}

func (conn *Conn) WaitOutMsg() <-chan *protocol.Message {
	return conn.outMsgChan
}

func (conn *Conn) WaitOutData() <-chan []byte {
	return conn.outBytesChan
}

//清空消息，避免有goroutine阻塞在 outMsgChan 或 outBytesChan
func (conn *Conn) Empty() {
	for {
		select {
		case msg := <-conn.outMsgChan:
			protocol.FreePoolMsg(msg)
		case <-conn.outBytesChan:

		default:
			if len(conn.outMsgChan) == 0 && len(conn.outBytesChan) == 0 {
				return
			}
		}
	}
}
