package server

import (
	"errors"
	"fmt"
	"github.com/kuhufu/cm/protocol"
	"github.com/kuhufu/cm/protocol/consts"
	"sync"
	"time"

	logger "github.com/kuhufu/cm/logger"
)

type Server struct {
	cm          *Manager
	opts        Options
	mu          sync.Mutex
	allChannels sync.Map //方便广播
	exitC       chan struct{}
}

func NewServer(opts ...Option) *Server {
	s := &Server{
		cm:    NewManager(),
		opts:  defaultOptions(),
		exitC: make(chan struct{}),
	}

	for _, opt := range opts {
		opt(&s.opts)
	}

	logger.Printf("auth_timeout: %v, heartbeat_timeout: %v", s.opts.AuthTimeout, s.opts.HeartbeatTimeout)

	return s
}

func (srv *Server) Close() error {
	close(srv.exitC)
	return nil
}

func (srv *Server) AddHandler(handler Handler) {
	if handler == nil {
		panic("message opts.Handler cannot be nil")
	}

	srv.opts.Handler = handler
}

func (srv *Server) optsCopy(opts ...Option) Options {
	srv.mu.Lock()
	optCpy := srv.opts
	srv.mu.Unlock()

	for _, opt := range opts {
		opt(&optCpy)
	}

	if optCpy.Handler == nil {
		panic("handler cannot be empty")
	}

	return optCpy
}

func (srv *Server) Run(addr string, opts ...Option) error {
	opt := srv.optsCopy(opts...)
	ln, err := getListener(addr, opt)
	if err != nil {
		return err
	}

	defer func() {
		ln.Close()
		logger.Infof("listener %v://%v exit", ln.Addr().Network(), ln.Addr().String())
	}()

	logger.Printf("listen on: %v", addr)

	network := ln.Addr().Network()
	for {
		if srv.exiting() {
			return errors.New("listener exit")
		}

		conn, err := ln.Accept()
		if err != nil {
			return err
		}

		logger.Printf("new connect: %v->%v", conn.RemoteAddr(), conn.LocalAddr())

		go func() {
			defer func() {
				if err := recover(); err != nil {
					logger.Error(err)
				}

			}()
			srv.serve(NewChannel(conn, network))
		}()
	}
}

func (srv *Server) exiting() bool {
	select {
	case <-srv.exitC:
		return true
	default:
	}

	return false
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
		logger.Println("auth timeout")
	})

	msg := protocol.NewMessage()
	for {
		if srv.exiting() {
			return
		}

		if _, err := msg.ReadFrom(channel.Conn); err != nil {
			logger.Println(err)
			return
		}

		logger.Debugf("receive message: %v", msg)

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
			err = fmt.Errorf("new connection must authentication %v", msg.Cmd())
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
			logger.Printf("%v, reader error: %v", channel, err)
		} else {
			logger.Printf("%v, reader exit", channel, channel.CreateTime)
		}
	}()

	heartbeatTimer = time.AfterFunc(srv.opts.HeartbeatTimeout, func() {
		channel.Close()
		logger.Println("first heartbeat timeout")
	})

	msg := protocol.NewMessage()
	for {
		if srv.exiting() {
			return
		}

		if _, err = msg.ReadFrom(channel.Conn); err != nil {
			return
		}

		logger.Debugf("channel_id: %v, msg: %s", channel.id, msg)

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
			err = fmt.Errorf("unkunown cmd: %v", msg.Cmd())
			return
		}
	}
}

func (srv *Server) writeLoop(channel *Channel) {
	var err error
	defer func() {
		if err != nil {
			logger.Printf("%v, writer error: %v", channel, err)
		} else {
			logger.Printf("%v, writer exit", channel)
		}
		channel.Close()
	}()

	for {
		if srv.exiting() {
			return
		}

		select {
		case <-srv.exitC:
			return
		case <-channel.Exit():
			return
		case msg := <-channel.WaitOutMsg():
			_, err = msg.WriteTo(channel)
			protocol.FreePoolMsg(msg)
			if err != nil {
				return
			}
		case data := <-channel.WaitOutBytes(): //多播专用chan
			if _, err = channel.Write(data); err != nil {
				return
			}
		}
	}
}

func (srv *Server) addChannel(channel *Channel, roomId string, channelId string) {
	if channelId == "" {
		panic("channel_id cannot be empty")
	}
	logger.Debugf("new channel, room_id: %v, channel_id: %v", roomId, channelId)

	srv.allChannels.Store(channel, nil)

	channel.Init(roomId, channelId)
	channel.OnClose = func() {
		logger.Debugf("channel onClose")

		// 1.从房间中移除
		srv.cm.GetOrCreate(roomId).DelIfEqual(channelId, channel)
		srv.allChannels.Delete(channel)

		//2.空房间移除
		if srv.cm.GetOrCreate(roomId).Size() == 0 {
			srv.cm.Del(roomId)
		}
	}

	oldChannel := srv.cm.GetOrCreate(roomId).AddOrReplace(channelId, channel)
	if oldChannel != nil {
		oldChannel.Close()
	}
}

//这里的单播，多播，广播的基本单位是room
func (srv *Server) Unicast(data []byte, roomId string, filters ...ChannelFilter) {
	var room *Room
	var ok bool

	if room, ok = srv.cm.Get(roomId); !ok {
		logger.Debugf("room:%v not exist", roomId)
	}

	data = srvPushMsgBytes(data)
	room.Broadcast(data, filters...)
}

func (srv *Server) Multicast(data []byte, roomIds []string, filters ...ChannelFilter) {
	data = srvPushMsgBytes(data)

	for _, id := range roomIds {
		if room, ok := srv.cm.Get(id); ok {
			room.Broadcast(data, filters...)
		}
	}
}

func (srv *Server) Broadcast(data []byte, filters ...ChannelFilter) {
	data = srvPushMsgBytes(data)

	srv.allChannels.Range(func(key, value interface{}) bool {
		c := key.(*Channel)
		for _, filter := range filters {
			if !filter(c) {
				return true
			}
		}
		c.EnterOutBytes(data)
		return true
	})
}

func (srv *Server) Range(f func(id string, room *Room) bool) {
	srv.cm.Range(f)
}

func (srv *Server) GetRoom(id string) (*Room, bool) {
	return srv.cm.Get(id)
}

func srvPushMsgBytes(data []byte) []byte {
	msg := protocol.GetPoolMsg().SetBody(data).SetCmd(consts.CmdServerPush)
	data = msg.Encode()
	protocol.FreePoolMsg(msg)

	return data
}
