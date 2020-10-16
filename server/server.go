package server

import (
	"fmt"
	"github.com/kuhufu/cm/protocol"
	"github.com/kuhufu/cm/protocol/consts"
	"sync"
	"time"

	logger "github.com/kuhufu/cm/logger"
	"github.com/kuhufu/cm/server/cm"
)

const DefaultAuthTimeout = time.Second * 10
const DefaultHeartBeatTimeout = time.Second * 90

type Handler interface {
	OnAuth(data []byte) *AuthReply
	OnReceive(srcConn *cm.Conn, data []byte) (resp []byte)
	OnClose(conn *cm.Conn)
}

type Server struct {
	authTimeout      time.Duration
	heartbeatTimeout time.Duration
	cm               *cm.ConnManager
	handler          Handler
	opts             Options
	mu               sync.Mutex
}

func NewServer(opts ...Option) *Server {
	s := &Server{
		cm:               cm.NewConnManager(),
		authTimeout:      DefaultAuthTimeout,
		heartbeatTimeout: DefaultHeartBeatTimeout,
	}

	for _, opt := range opts {
		opt(s)
	}

	if s.handler == nil {
		panic("message handler cannot be nil")
	}

	logger.Printf("auth_timeout: %v, heartbeat_timeout: %v", s.authTimeout, s.heartbeatTimeout)

	return s
}

func (srv *Server) optsCopy(opts ...Option) Options {
	srvCpy := *srv

	srv.mu.Lock()
	defer srv.mu.Unlock()

	for _, opt := range opts {
		opt(&srvCpy)
	}

	return srvCpy.opts
}

func (srv *Server) Run(addr string, opts ...Option) error {
	ln, err := GetListener(addr, srv.optsCopy(opts...))
	if err != nil {
		return err
	}

	logger.Printf("listen on: %v", addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		logger.Printf("new connect: %v", conn.RemoteAddr())
		go srv.serve(cm.NewConn(conn))
	}
}

func (srv *Server) serve(conn *cm.Conn) {
	var err error
	defer func() {
		if err != nil {
			logger.Println(err)
		}
		conn.Close()
	}()

	go srv.writeLoop(conn)

	AuthTimer := time.AfterFunc(srv.authTimeout, func() {
		conn.Close()
		logger.Println("认证超时")
	})

	msg := protocol.NewMessage()
	for {
		if _, err := msg.ReadFrom(conn); err != nil {
			logger.Println(err)
			return
		}

		logger.Debug(msg)

		switch msg.Cmd() {
		case consts.CmdAuth:
			reply := srv.handler.OnAuth(msg.Body())

			if err = reply.err; err != nil {
				return
			}

			conn.EnterOutMsg(CreateReplyMessage(msg, reply.Data))

			if reply.Ok {
				if !AuthTimer.Stop() {
					err = ErrAuthTimeout
					return
				}

				//为连接添加拓展信息
				for k, v := range reply.Metadata {
					conn.Metadata.Store(k, v)
				}

				srv.addConn(conn, reply.UserId, reply.ConnId, reply.GroupIds)
				goto authOk
			}
		default:
			err = fmt.Errorf("新连接需要认证 %v", msg.Cmd())
			return
		}
	}
authOk:
	srv.readLoop(conn)
}

func (srv *Server) readLoop(conn *cm.Conn) {
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

		srv.handler.OnClose(conn)
	}()

	heartbeatTimer = time.AfterFunc(srv.heartbeatTimeout, func() {
		conn.Close()
		logger.Println("第一个心跳超时")
	})

	msg := protocol.NewMessage()
	for {
		logger.Debugf("connId: %v, msg: %s", conn.Id, msg)

		if _, err = msg.ReadFrom(conn); err != nil {
			return
		}

		switch msg.Cmd() {
		case consts.CmdPush:
			data := srv.handler.OnReceive(conn, msg.Body())
			conn.EnterOutMsg(CreateReplyMessage(msg, data))
		case consts.CmdHeartbeat:
			if !heartbeatTimer.Stop() {
				err = ErrHeartbeatTimeout
				return
			}
			heartbeatTimer.Reset(srv.heartbeatTimeout)
			conn.EnterOutMsg(CreateReplyMessage(msg, nil))
		case consts.CmdClose:
			return
		default:
			err = fmt.Errorf("未知的cmd: %v", msg.Cmd())
			return
		}
	}
}

func (srv *Server) writeLoop(conn *cm.Conn) {
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
			_, err = msg.WriteTo(conn)
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
	conn, ok := srv.cm.GetConn(connId)
	if !ok {
		logger.Printf("连接不存在或已关闭, connId: %v", connId)
		return ErrConnNotExist
	}

	msg := protocol.GetPoolMsg().SetBody(data).SetCmd(consts.CmdServerPush)
	conn.EnterOutMsg(msg)
	protocol.FreePoolMsg(msg)

	return nil
}

func (srv *Server) PushToUser(uid string, data []byte) {
	group, ok := srv.cm.GetDeviceGroup(uid)
	if !ok || group.Size() == 0 {
		return
	}

	data = srvPushMsgBytes(data)

	group.ForEach(func(conn *cm.Conn) {
		conn.EnterOutBytes(data)
	})

}

//推送到设备组
func (srv *Server) PushToDeviceGroup(userId string, data []byte) error {
	data = srvPushMsgBytes(data)

	group, ok := srv.cm.GetDeviceGroup(userId)
	if !ok {
		logger.Printf("连接不存在或已关闭, userId: %v", userId)
		return ErrConnNotExist
	}

	group.ForEach(func(conn *cm.Conn) {
		conn.EnterOutBytes(data)
	})

	return nil
}

//从群组中移除
func (srv *Server) RemoveFromGroup(groupId, userId string) {
	srv.cm.RemoveFromGroup(userId, groupId)
}

//推送到群组
func (srv *Server) PushToGroup(groupId string, data []byte) {
	g, ok := srv.cm.GetGroup(groupId)
	if !ok {
		return
	}

	data = srvPushMsgBytes(data)

	g.ForEach(func(id cm.UserId, g *cm.DeviceGroup) {
		g.ForEach(func(conn *cm.Conn) {
			conn.EnterOutBytes(data)
		})
	})
}

func (srv *Server) addConn(conn *cm.Conn, userId, connId string, groupIds []string) {
	if connId == "" {
		panic("connId cannot be empty")
	}

	conn.Init(userId, connId)

	logger.Printf("newConn: %v:%v", conn.Id, conn.Version)

	var oldConn *cm.Conn
	srv.cm.With(func() {
		oldConn = srv.cm.AddOrReplaceNoSync(connId, conn)
		srv.cm.AddToGroupNoSync(conn.UserId, groupIds)
	})

	if oldConn != nil {
		oldConn.Close()
	}
}

func (srv *Server) unicast(data []byte, connId string) error {
	conn, ok := srv.cm.GetConn(connId)
	if !ok {
		logger.Printf("连接不存在或已关闭, connId: %v", connId)
		return ErrConnNotExist
	}

	msg := protocol.GetPoolMsg().SetBody(data).SetCmd(consts.CmdServerPush)
	conn.EnterOutMsg(msg)

	return nil
}

func (srv *Server) multicast(data []byte, connIds ...string) {
	data = srvPushMsgBytes(data)

	items := make([]*cm.Conn, 0, len(connIds))

	for _, connId := range connIds {
		conn, _ := srv.cm.GetConn(connId)
		items = append(items, conn)
		conn.EnterOutBytes(data)
	}
}

func (srv *Server) broadcast(data []byte) {
	data = srvPushMsgBytes(data)

	all := srv.cm.AllConn()
	for _, conn := range all {
		conn.EnterOutBytes(data)
	}
}

func srvPushMsgBytes(data []byte) []byte {
	msg := protocol.GetPoolMsg().SetBody(data).SetCmd(consts.CmdServerPush)
	data = msg.Encode()
	protocol.FreePoolMsg(msg)

	return data
}
