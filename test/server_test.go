package test

import (
	"encoding/json"
	"fmt"
	"github.com/kuhufu/cm/server"
	"net/http"
	_ "net/http/pprof"
	"net/url"
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
	srv := server.NewServer(
		server.WithHandler(&Handler{}),
		server.WithAuthTimeout(time.Second*1000),
		server.WithHeartbeatTimeout(time.Second*300),
		server.WithDebugLog(),
		server.WithCertAndKeyFile("cert/cert.pem", "cert/key.pem"),
	)

	go func() {
		err := srv.Run("wss://0.0.0.0:8081/ws")
		if err != nil {
			t.Error(err)
		}
	}()

	go func() {
		err := srv.Run("tcp://0.0.0.0:8080")
		if err != nil {
			t.Error(err)
		}
	}()

	go func() {
		for {
			time.Sleep(time.Second)
			//err := srv.Unicast([]byte("hello"), "1")
			bytes := make([]byte, 40960)
			srv.Broadcast(bytes, func(channel *server.Channel) bool {
				return channel.Id() == "web"
			})
		}
	}()

	go func() {
		if err := http.ListenAndServe(":8888", nil); err != nil {
			panic(err)
		}
	}()

	select {}
}

func TestScheme(t *testing.T) {
	parse, err := url.Parse("localhost:8080")
	if err != nil {
		t.Error(err)
	}

	fmt.Println(parse.Scheme)
	fmt.Println(parse.Host)
	fmt.Println(parse.Hostname())
	fmt.Println(parse.Path)
	fmt.Println(parse.RequestURI())
}
