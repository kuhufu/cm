package server

import (
	"fmt"
	"github.com/kuhufu/cm/protocol"
	"net"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type Conn struct {
	net.Conn
	Id             string
	GroupIds       []string
	outMsgChan     chan *protocol.Message
	exitChan       chan struct{}
	writerExitChan chan struct{}
	readerExitChan chan struct{}

	closeOnce      sync.Once //保证连接只关闭一次
	AuthTime       time.Time
	version        string   //创建连接时的版本号
	closedByServer uint32   //辅助 server.CloseConn 函数， 保证每个conn只被server.CloseConn使用一次
	Extends        sync.Map //拓展信息可自由添加

	sync.Mutex
}

func NewConn(conn net.Conn) *Conn {
	c := &Conn{
		Conn:           conn,
		exitChan:       make(chan struct{}),
		outMsgChan:     make(chan *protocol.Message, 4),
		closeOnce:      sync.Once{},
		writerExitChan: make(chan struct{}),
		readerExitChan: make(chan struct{}),
		Extends:        sync.Map{},
	}

	//创建时间+内存地址，本地测试中创建时间可能会相同。在创建时间相同的情况下，内存地址必不相同
	c.version = fmt.Sprintf("%x-%x", time.Now().UnixNano(), uintptr(unsafe.Pointer(c)))

	return c
}

func (conn *Conn) Close() error {
	var err error
	conn.closeOnce.Do(func() {
		close(conn.exitChan)
		err = conn.Conn.Close()
		conn.EmptyMsg()
	})
	return err
}

func (conn *Conn) Exit() <-chan struct{} {
	return conn.exitChan
}

func (conn *Conn) WriterExit() {
	close(conn.writerExitChan)
}

func (conn *Conn) ReaderExit() {
	close(conn.readerExitChan)
}

//等待完全退出（完全退出指reader和writer均退出）
func (conn *Conn) WaitFullExit() {
	<-conn.readerExitChan
	<-conn.writerExitChan
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

func (conn *Conn) WaitOutMsg() <-chan *protocol.Message {
	return conn.outMsgChan
}

//清空消息
func (conn *Conn) EmptyMsg() {
	for {
		select {
		case msg := <-conn.outMsgChan:
			protocol.FreePoolMsg(msg)
		default:
			if len(conn.outMsgChan) == 0 {
				return
			}
		}
	}
}

// 标记是否被server.CloseConn第一次调用
func (conn *Conn) FirstCloseByServer() bool {
	return atomic.CompareAndSwapUint32(&conn.closedByServer, 0, 1)
}
