package test

import (
	bytes2 "bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/kuhufu/cm/protocol"
	"github.com/kuhufu/cm/protocol/consts"
	"github.com/kuhufu/cm/server"
	"github.com/kuhufu/cm/server/cm"
	"io/ioutil"
	"log"
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
		Ok:       true,
		ConnId:   f.Uid + ":" + f.Os,
		UserId:   f.Uid,
		GroupIds: []string{"g1", "g2"},
		Data:     []byte("hello"),
	}
}

func (h Handler) OnReceive(srcConn *cm.Conn, data []byte) (resp []byte) {
	return data
}

func (h Handler) OnClose(conn *cm.Conn) {
	fmt.Println("连接已关闭")
}

func Test_Server(t *testing.T) {
	srv := server.NewServer(
		server.WithMessageHandler(&Handler{}),
		server.WithAuthTimeout(time.Second*10),
		server.WithHeartbeatTimeout(time.Minute*100),
		server.WithCertAndKeyFile("cert.pem", "key.pem"),
	)

	go func() {
		for {
			time.Sleep(time.Second)
			srv.PushToDeviceGroup("1", []byte("hello"))
		}
	}()

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

	select {}
}

func Test_ClientWs(t *testing.T) {
	//development 123.56.103.77:7090
	//production kfws.qiyejiaoyou.com:7090
	f := func(uid string, os string) {
		dialer := websocket.Dialer{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		conn, response, err := dialer.Dial("wss://localhost:8081/ws", nil)
		if err != nil {
			t.Error(err)
			return
		}

		fmt.Println(conn.RemoteAddr())

		bytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			t.Error(err)
		}

		fmt.Println(string(bytes))

		time.Sleep(time.Millisecond)

		msg := protocol.NewDefaultMessage()
		msg.SetCmd(consts.CmdAuth)
		msg.SetBody([]byte(fmt.Sprintf(`{"uid":"%v","os":"%v"}`, uid, os)))

		fmt.Println(msg)

		err = conn.WriteMessage(websocket.BinaryMessage, msg.Encode())
		if err != nil {
			t.Error(err)
			return
		}

		for {
			//读
			_, data, err := conn.ReadMessage()
			if err != nil {
				log.Println()
				t.Error(err)
				return
			}

			msg.ReadFrom(bytes2.NewReader(data))
			log.Println(os+":receive:", msg)
		}
	}

	go f("1", "web")

	time.Sleep(time.Hour)
}

func Test_ClientTcp(t *testing.T) {
	//development 123.56.103.77:7090
	//production kfws.qiyejiaoyou.com:7090
	f := func(uid string, os string) {
		conn, err := tls.Dial("tcp", "localhost:8080", &tls.Config{InsecureSkipVerify: true})
		if err != nil {
			t.Error(err)
			return
		}

		fmt.Println(conn.RemoteAddr())

		time.Sleep(time.Millisecond)

		msg := protocol.NewDefaultMessage()
		msg.SetCmd(consts.CmdAuth)
		msg.SetBody([]byte(fmt.Sprintf(`{"uid":"%v","os":"%v"}`, uid, os)))

		fmt.Println(msg)

		_, err = conn.Write(msg.Encode())
		if err != nil {
			t.Error(err)
			return
		}

		for {
			//读
			msg.ReadFrom(conn)
			log.Println(os+":receive:", msg)
		}
	}

	go f("1", "android")

	time.Sleep(time.Hour)
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
