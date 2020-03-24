package server

import (
	"errors"
	"fmt"
	"sync"
	"time"

	logger "github.com/kuhufu/cm/logger"
	"github.com/kuhufu/cm/protocol"
	"github.com/kuhufu/cm/server/cm"
	"github.com/kuhufu/cm/server/listener"
)

const DefaultAuthTimeout = time.Second * 10
const DefaultHeartBeatTimeout = time.Second * 90

type MessageHandler interface {
	Auth(data []byte) *AuthReply
	PushIn(srcConn *cm.Conn, data []byte) (resp []byte)
	OnConnClose(conn *cm.Conn)
}

type Server struct {
	AuthTimeout      time.Duration
	HeartbeatTimeout time.Duration

	cm             *cm.ConnManager
	addr           string
	messageHandler MessageHandler

	connMux sync.Mutex
}

func NewServer(opts ...Option) *Server {
	s := &Server{
		cm:               cm.NewConnManager(),
		AuthTimeout:      DefaultAuthTimeout,
		HeartbeatTimeout: DefaultHeartBeatTimeout,
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
		go srv.Handle(cm.NewConn(conn))
	}
}

func (srv *Server) Handle(conn *cm.Conn) {
	var err error
	defer func() {
		if err != nil {
			logger.Println(err)
		}
		conn.Close()
	}()

	go srv.HandleWriter(conn)

	AuthTimer := time.AfterFunc(srv.AuthTimeout, func() {
		conn.Close()
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

			if err = reply.err; err != nil {
				return
			}

			conn.EnterOutMsg(CreateReplyMessage(msg, reply.Data))

			if reply.Ok {
				if !AuthTimer.Stop() {
					err = errors.New("认证超时")
					return
				}

				//为连接添加拓展信息
				for k, v := range reply.Metadata {
					conn.Metadata.Store(k, v)
				}

				srv.AddConn(conn, reply.UserId, reply.ConnId, reply.GroupIds)
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

func (srv *Server) HandleReader(conn *cm.Conn) {
	var err error
	var heartbeatTimer *time.Timer

	defer func() { //在defer里面关闭连接
		srv.cm.RemoveConn(conn)
		heartbeatTimer.Stop()

		if err != nil {
			logger.Printf("connId: %v:%v, reader出错: %v", conn.Id, conn.Version, err)
		} else {
			logger.Printf("connId: %v:%v, reader退出", conn.Id, conn.Version)
		}
	}()

	heartbeatTimer = time.AfterFunc(srv.HeartbeatTimeout, func() {
		conn.Close()
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

func (srv *Server) HandleWriter(conn *cm.Conn) {
	var err error
	defer func() {
		srv.cm.RemoveConn(conn)
		if err != nil {
			logger.Printf("connId: %v:%v, writer出错: %v", conn.Id, conn.Version, err)
		} else {
			logger.Printf("connId: %v:%v, writer退出", conn.Id, conn.Version)
		}
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
		case data := <-conn.WaitOutBytes(): //多播专用chan
			if _, err = conn.Write(data); err != nil {
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

//推送到设备组
func (srv *Server) PushToDeviceGroup(userId string, data []byte) error {
	msg := protocol.GetPoolMsg().SetBody(data).SetCmd(protocol.CmdServerPush)
	data = msg.Encode()
	protocol.FreePoolMsg(msg)

	if group, ok := srv.cm.GetDeviceGroup(userId); ok {
		group.ForEach(func(conn *cm.Conn) {
			conn.EnterOutBytes(data)
		})
	} else {
		logger.Printf("连接不存在或已关闭, userId: %v", userId)
		return ErrConnNotExist
	}

	return nil
}

//推送到群组
func (srv *Server) PushToGroup(groupId string, data []byte) {
	msg := protocol.GetPoolMsg().SetBody(data).SetCmd(protocol.CmdServerPush)
	data = msg.Encode()
	protocol.FreePoolMsg(msg)

	if g, ok := srv.cm.GetGroup(groupId); ok {
		g.ForEach(func(id cm.UserId, g *cm.DeviceGroup) {
			g.ForEach(func(conn *cm.Conn) {
				conn.EnterOutBytes(data)
			})
		})
	}
}

func (srv *Server) AddConn(conn *cm.Conn, userId, connId string, groupIds []string) {
	if connId == "" {
		panic("connId cannot be empty")
	}

	conn.Init(userId, connId)

	logger.Printf("newConn: %v:%v", conn.Id, conn.Version)

	var oldConn *cm.Conn
	srv.cm.With(func() {
		oldConn = srv.cm.AddOrReplaceSyncNo(connId, conn)
		srv.cm.AddToGroupSyncNo(conn.UserId, groupIds)
	})

	if oldConn != nil {
		oldConn.Close()
	}
}

func (srv *Server) GetConn(connId string) (*cm.Conn, bool) {
	conn, ok := srv.cm.GetConn(connId)
	if !ok {
		return nil, false
	}

	return conn, true
}
