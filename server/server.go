package server

import (
	"errors"
	"fmt"
	logger "github.com/kuhufu/cm/logger"
	"github.com/kuhufu/cm/protocol"
	"github.com/kuhufu/cm/server/listener"
	"sync"
	"time"
)

const DefAuthTimeout = time.Second * 10
const DefHeartBeatTimeout = time.Second * 90

type MessageHandler interface {
	Auth(data []byte) *AuthReply
	PushIn(srcConn *Conn, data []byte) (resp []byte)
	OnConnClose(conn *Conn)
}

type Server struct {
	AuthTimeout      time.Duration
	HeartbeatTimeout time.Duration

	globalConnMap  *sync.Map //全局连接map
	connGroupMap   *sync.Map //连接组map
	addr           string
	messageHandler MessageHandler

	connMux sync.Mutex
}

func NewServer(opts ...optFunc) *Server {
	s := &Server{
		globalConnMap:    &sync.Map{},
		connGroupMap:     &sync.Map{},
		AuthTimeout:      DefAuthTimeout,
		HeartbeatTimeout: DefHeartBeatTimeout,
	}

	for _, opt := range opts {
		opt(s)
	}

	if s.messageHandler == nil {
		panic("message handler cannot be nil")
	}

	logger.Printf("auth_timeout: %v, heartbeat_timeout: %v", s.AuthTimeout, s.HeartbeatTimeout)

	return s
}

func (srv *Server) Run(network, addr string) error {
	ln, err := listener.Get(network, addr)
	if err != nil {
		return err
	}

	logger.Printf("listen on: %v://%v", network, addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		logger.Printf("new connect: %v", conn.RemoteAddr())
		go srv.Handle(NewConn(conn))
	}
}

func (srv *Server) Handle(conn *Conn) {
	var err error
	defer func() {
		if err != nil {
			logger.Println(err)
		}
		conn.Close()
	}()

	go srv.HandleWriter(conn)

	AuthTimer := time.AfterFunc(srv.AuthTimeout, func() {
		srv.CloseConn(conn)
		logger.Println("认证超时")
	})

	msg := protocol.NewMessage()
	for {
		if err := msg.Decode(conn); err != nil {
			logger.Println(err)
			return
		}

		logger.Debug(msg)

		switch msg.Cmd() {
		case protocol.CmdAuth:
			reply := srv.messageHandler.Auth(msg.Body())

			authOk, authTime, connId, groupIds, data, extends := reply.Ok, reply.AuthTime, reply.ConnId, reply.GroupIds, reply.Data, reply.Extends

			if err = reply.err; err != nil {
				return
			}

			conn.EnterOutMsg(CreateReplyMessage(msg, data))

			if authOk {
				if !AuthTimer.Stop() {
					err = errors.New("认证超时")
					return
				}

				//为连接添加拓展信息
				for k, v := range extends {
					conn.Extends.Store(k, v)
				}

				srv.AddConn(conn, connId, groupIds, authTime)
				goto authOk
			}
		default:
			err = fmt.Errorf("新连接需要认证 %v", msg.Cmd())
			return
		}
	}
authOk:
	srv.HandleReader(conn)
}

func (srv *Server) HandleReader(conn *Conn) {
	var err error
	var heartbeatTimer *time.Timer

	defer func() { //在defer里面关闭连接
		srv.CloseConn(conn)
		heartbeatTimer.Stop()
		if err != nil {
			logger.Printf("connId: %v:%v, reader出错: %v", conn.Id, conn.version, err)
		} else {
			logger.Printf("connId: %v:%v, reader退出", conn.Id, conn.version)
		}

		conn.ReaderExit()
	}()

	heartbeatTimer = time.AfterFunc(srv.HeartbeatTimeout, func() {
		srv.CloseConn(conn)
		logger.Println("第一个心跳超时")
	})

	msg := protocol.NewMessage()
	for {
		err = msg.Decode(conn)

		logger.Debugf("connId: %v, msg: %s", conn.Id, msg)

		if err != nil {
			return
		}

		switch msg.Cmd() {
		case protocol.CmdPush:
			data := srv.messageHandler.PushIn(conn, msg.Body())
			conn.EnterOutMsg(CreateReplyMessage(msg, data))
		case protocol.CmdHeartbeat:
			if !heartbeatTimer.Reset(srv.HeartbeatTimeout) {
				err = errors.New("心跳超时")
				return
			}
			conn.EnterOutMsg(CreateReplyMessage(msg, nil))
		case protocol.CmdClose:
			return
		default:
			err = fmt.Errorf("未知的cmd: %v", msg.Cmd())
			return
		}
	}
}

func (srv *Server) HandleWriter(conn *Conn) {
	var err error

	defer func() {
		if err != nil {
			logger.Printf("connId: %v:%v, writer出错: %v", conn.Id, conn.version, err)
		} else {
			logger.Printf("connId: %v:%v, writer退出", conn.Id, conn.version)
		}
		srv.CloseConn(conn)
		conn.WriterExit()
	}()

	for {
		select {
		case <-conn.Exit():
			return
		case msg := <-conn.WaitOutMsg():
			err = msg.WriteTo(conn)
			//不能对这个连接进行并发写，WriteTo操作不是原子的，WriteTo先写header再写body，如果对连接进行并发写，会出现错误的数据
			//正常情况下header和body一一对应[header1,body1,header2,body2]，并发写可能会出现[header1,header2,body1,body2]
			protocol.FreePoolMsg(msg)
			if err != nil {
				return
			}
		}
	}
}

//推送到连接
func (srv *Server) PushToConn(connId string, data []byte) error {
	msg := protocol.GetPoolMsg()
	msg.SetBody(data).SetCmd(protocol.CmdServerPush)
	if conn, ok := srv.GetConn(connId); ok {
		conn.EnterOutMsg(msg)
	} else {
		logger.Printf("连接不存在或已关闭, connId: %v", connId)
		protocol.FreePoolMsg(msg)
		return ErrConnNotExist
	}

	return nil
}

//推送到连接组
func (srv *Server) PushToGroup(groupId string, data []byte) {
	if g, ok := srv.connGroupMap.Load(groupId); ok {
		g.(*sync.Map).Range(func(key, value interface{}) bool {
			srv.PushToConn(key.(string), data)
			return true
		})
	}
}

func (srv *Server) AddConn(conn *Conn, connId string, groupIds []string, authTime time.Time) {
	if connId == "" {
		panic("connId cannot be empty")
	}

	conn.Id = connId
	conn.AuthTime = authTime

	logger.Printf("newConn: %v:%v", conn.Id, conn.version)

	srv.connMux.Lock()
	if oldConn, exist := srv.GetConn(connId); exist {
		logger.Printf("connId: %v:%v has same connId with new conn, wait to close it", oldConn.Id, oldConn.version)
		oldConn.Close()
		oldConn.WaitFullExit()
		logger.Printf("connId: %v:%v closedByServer, because new conn has same connId with it", oldConn.Id, oldConn.version)
	}
	srv.globalConnMap.Store(connId, conn)
	srv.connMux.Unlock()

	conn.GroupIds = append(conn.GroupIds, groupIds...)
	for _, groupId := range groupIds {
		load, _ := srv.connGroupMap.LoadOrStore(groupId, &sync.Map{})
		load.(*sync.Map).Store(connId, conn)
	}
}

func (srv *Server) GetConn(connId string) (*Conn, bool) {
	value, ok := srv.globalConnMap.Load(connId)
	if !ok {
		return nil, false
	}

	return value.(*Conn), true
}

//从conn map和group中移除conn，并关闭conn
func (srv *Server) CloseConn(conn *Conn) {
	if !conn.FirstCloseByServer() {
		//不成功就说明已经对该conn调用过CloseConn方法
		logger.Debugf("connId: %v:%v, 已调用过CloseConn方法", conn.Id, conn.version)
		return
	}

	if curConn, ok := srv.globalConnMap.Load(conn.Id); !ok || curConn.(*Conn).version != conn.version {
		return
	}
	srv.globalConnMap.Delete(conn.Id)
	for _, groupId := range conn.GroupIds {
		if load, ok := srv.connGroupMap.Load(groupId); ok {
			load.(*sync.Map).Delete(conn.Id)
		}
	}

	conn.Close()
	srv.messageHandler.OnConnClose(conn)
}
