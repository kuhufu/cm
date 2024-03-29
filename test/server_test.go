package test

import (
	"encoding/json"
	"fmt"
	"github.com/kuhufu/cm/protocol"
	"github.com/kuhufu/cm/server"
	_ "net/http/pprof"
	"testing"
	"time"
)

type Handler struct {
}

func (h Handler) OnAuth(data []byte) *server.AuthReply {

	f := struct {
		Os  string `json:"os"`
		Uid string `json:"uid"`
	}{}

	json.Unmarshal(data, &f)

	return &server.AuthReply{
		Ok:        true,
		ChannelId: f.Os,
		RoomId:    f.Uid,
		Data:      []byte("hello"),
	}
}

func (h Handler) OnReceive(c *server.Channel, data []byte) (resp []byte) {
	fmt.Println("OnReceive")
	return data
}

func (h Handler) OnClose(conn *server.Channel) {
	fmt.Println("OnClose 连接已关闭")
}

func Test_Server(t *testing.T) {
	binarySrv := server.NewServer(
		server.WithHandler(&Handler{}),
		server.WithAuthTimeout(time.Second*1000),
		server.WithHeartbeatTimeout(time.Second*300),
		server.WithDebugLog(),
		server.WithCertAndKeyFile("cert/cert.pem", "cert/key.pem"),
		server.WithMsgProtocol(protocol.BINARY),
	)

	go func() {
		err := binarySrv.Run("tcp://0.0.0.0:8080")
		if err != nil {
			t.Error(err)
		}
	}()

	jsonSrv := server.NewServer(
		server.WithHandler(&Handler{}),
		server.WithAuthTimeout(time.Second*1000),
		server.WithHeartbeatTimeout(time.Second*300),
		server.WithDebugLog(),
		server.WithCertAndKeyFile("cert/cert.pem", "cert/key.pem"),
		server.WithMsgProtocol(protocol.JSON),
	)

	go func() {
		err := jsonSrv.Run("wss://0.0.0.0:8081/ws")
		if err != nil {
			t.Error(err)
		}
	}()

	go func() {
		for {
			time.Sleep(time.Second)
			bytes := []byte("hello world")

			binarySrv.Broadcast(bytes, func(channel *server.Channel) bool {
				return channel.Id() == "web"
			})

			jsonSrv.Broadcast(bytes, func(channel *server.Channel) bool {
				return channel.Id() == "web"
			})
		}
	}()

	select {}
}
