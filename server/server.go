package server

import (
	"fmt"
	"github.com/kuhufu/cm/protocol"
	"github.com/kuhufu/cm/protocol/consts"
	"sync"
	"time"

	logger "github.com/kuhufu/cm/logger"
)

const DefaultAuthTimeout = time.Second * 10
const DefaultHeartBeatTimeout = time.Second * 90

type Handler interface {
	OnAuth(data []byte) *AuthReply
	OnReceive(channel *Channel, data []byte) (resp []byte)
	OnClose(channel *Channel)
}

type Server struct {
	cm          *Manager
	opts        Options
	mu          sync.Mutex
	allChannels sync.Map //方便广播
}

func NewServer(opts ...Option) *Server {
	s := &Server{
		cm: NewManager(),
		opts: Options{
			AuthTimeout:      DefaultAuthTimeout,
			HeartbeatTimeout: DefaultAuthTimeout,
		},
	}

	for _, opt := range opts {
		opt(&s.opts)
	}

	if s.opts.Handler == nil {
		panic("message opts.Handler cannot be nil")
	}

	logger.Printf("auth_timeout: %v, heartbeat_timeout: %v", s.opts.AuthTimeout, s.opts.HeartbeatTimeout)

	return s
}

func (srv *Server) optsCopy(opts ...Option) Options {
	srv.mu.Lock()
	optCpy := srv.opts
	srv.mu.Unlock()

	for _, opt := range opts {
		opt(&optCpy)
	}

	return optCpy
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
		go srv.serve(&Channel{
			Conn: NewConn(conn),
		})
	}
}

func (srv *Server) serve(channel *Channel) {
	var err error
	defer func() {
		if err != nil {
			logger.Println(err)
		}
		channel.Close()
		srv.opts.Handler.OnClose(channel)
	}()

	go srv.writeLoop(channel)

	AuthTimer := time.AfterFunc(srv.opts.AuthTimeout, func() {
		channel.Close()
		logger.Println("认证超时")
	})

	msg := protocol.NewMessage()
	for {
		if _, err := msg.ReadFrom(channel); err != nil {
			logger.Println(err)
			return
		}

		//logger.Debug(msg)

		switch msg.Cmd() {
		case consts.CmdAuth:
			reply := srv.opts.Handler.OnAuth(msg.Body())

			if err = reply.err; err != nil {
				return
			}

			channel.EnterOutMsg(CreateReplyMessage(msg, reply.Data))

			if reply.Ok {
				if !AuthTimer.Stop() {
					err = ErrAuthTimeout
					return
				}

				//为连接添加拓展信息
				for k, v := range reply.Metadata {
					channel.Metadata.Store(k, v)
				}

				srv.addChannel(channel, reply.RoomId, reply.ChannelId)
				goto authOk
			}
		default:
			err = fmt.Errorf("新连接需要认证 %v", msg.Cmd())
			return
		}
	}
authOk:
	srv.readLoop(channel)
}

func (srv *Server) readLoop(channel *Channel) {
	var err error
	var heartbeatTimer *time.Timer

	defer func() { //在defer里面关闭连接
		heartbeatTimer.Stop()
		if err != nil {
			logger.Printf("%v, reader 出错: %v", channel, err)
		} else {
			logger.Printf("%v, reader 退出", channel, channel.CreateTime)
		}
	}()

	heartbeatTimer = time.AfterFunc(srv.opts.HeartbeatTimeout, func() {
		channel.Close()
		logger.Println("第一个心跳超时")
	})

	msg := protocol.NewMessage()
	for {
		if _, err = msg.ReadFrom(channel); err != nil {
			return
		}

		logger.Debugf("channel_id: %v, msg: %s", channel.Id, msg)

		switch msg.Cmd() {
		case consts.CmdPush:
			data := srv.opts.Handler.OnReceive(channel, msg.Body())
			channel.EnterOutMsg(CreateReplyMessage(msg, data))
		case consts.CmdHeartbeat:
			if !heartbeatTimer.Stop() {
				err = ErrHeartbeatTimeout
				return
			}
			heartbeatTimer.Reset(srv.opts.HeartbeatTimeout)
			channel.EnterOutMsg(CreateReplyMessage(msg, nil))
		case consts.CmdClose:
			return
		default:
			err = fmt.Errorf("未知的cmd: %v", msg.Cmd())
			return
		}
	}
}

func (srv *Server) writeLoop(conn *Channel) {
	var err error
	defer func() {
		if err != nil {
			logger.Printf("%v, writer 出错: %v", conn, err)
		} else {
			logger.Printf("%v, writer 退出", conn)
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

func (srv *Server) addChannel(channel *Channel, roomId RoomId, channelId ChannelId) {
	if channelId == "" {
		panic("channelId cannot be empty")
	}
	logger.Debugf("new channel, room_id: %v, channel_id: %v", roomId, channelId)

	srv.allChannels.Store(channel, nil)

	channel.Id = channelId
	channel.ClientType = channelId
	channel.OnClose = func() {
		logger.Debugf("channel onClose")
		srv.cm.GetOrCreate(roomId).DelIfEqual(channelId, channel)
		srv.allChannels.Delete(channel)
	}

	oldChannel := srv.cm.GetOrCreate(roomId).AddOrReplace(channelId, channel)
	if oldChannel != nil {
		oldChannel.Close()
	}
}

func (srv *Server) Unicast(data []byte, channelId RoomId) error {
	var channel *Room
	var ok bool

	if channel, ok = srv.cm.Get(channelId); !ok {
		return ErrConnNotExist
	}

	data = srvPushMsgBytes(data)
	channel.Range(func(key ClientType, conn *Channel) {
		conn.EnterOutBytes(data)
	})

	return nil
}

func (srv *Server) Multicast(data []byte, roomIds ...RoomId) {
	data = srvPushMsgBytes(data)

	for _, id := range roomIds {
		if channel, ok := srv.cm.Get(id); ok {
			channel.Range(func(id ChannelId, conn *Channel) {
				conn.EnterOutBytes(data)
			})
		}
	}
}

func (srv *Server) Broadcast(data []byte) {
	data = srvPushMsgBytes(data)

	srv.allChannels.Range(func(key, value interface{}) bool {
		key.(*Channel).EnterOutBytes(data)
		return true
	})
}

func srvPushMsgBytes(data []byte) []byte {
	msg := protocol.GetPoolMsg().SetBody(data).SetCmd(consts.CmdServerPush)
	data = msg.Encode()
	protocol.FreePoolMsg(msg)

	return data
}
