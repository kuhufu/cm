package server

import (
	bytes2 "bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/kuhufu/cm/protocol/binary"
	"github.com/kuhufu/cm/server/cm"
	"io/ioutil"
	"log"
	"net"
	"testing"
	"time"
)

type Handler struct {
}

func (h Handler) OnAuth(data []byte) *AuthReply {

	f := struct {
		Os  string `json:"os"`
		Uid string `json:"uid"`
	}{}

	json.Unmarshal(data, &f)

	return &AuthReply{
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

func (h Handler) OnConnClose(conn *cm.Conn) {
	fmt.Println("连接已关闭")
}

func Test_Server(t *testing.T) {
	srv := NewServer(
		WithMessageHandler(&Handler{}),
		WithAuthTimeout(time.Second*10),
		WithHeartbeatTimeout(time.Minute*100),
	)

	go func() {
		for {
			time.Sleep(time.Second)
			srv.PushToDeviceGroup("1", []byte("hello"))
		}
	}()

	go func() {
		err := srv.Run("ws", "0.0.0.0:8081")
		if err != nil {
			t.Error(err)
		}
	}()

	go func() {
		err := srv.Run("tcp", "0.0.0.0:8080")
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
		conn, response, err := websocket.DefaultDialer.Dial("ws://localhost:8081/ws", nil)
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

		msg := binary.NewDefaultMessage()
		msg.SetCmd(binary.CmdAuth)
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

			msg.Decode(bytes2.NewReader(data))
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
		conn, err := net.Dial("tcp", "localhost:8080")
		if err != nil {
			t.Error(err)
			return
		}

		fmt.Println(conn.RemoteAddr())

		time.Sleep(time.Millisecond)

		msg := binary.NewDefaultMessage()
		msg.SetCmd(binary.CmdAuth)
		msg.SetBody([]byte(fmt.Sprintf(`{"uid":"%v","os":"%v"}`, uid, os)))

		fmt.Println(msg)

		_, err = conn.Write(msg.Encode())
		if err != nil {
			t.Error(err)
			return
		}

		for {
			//读
			msg.Decode(conn)
			log.Println(os+":receive:", msg)
		}
	}

	go f("1", "android")

	time.Sleep(time.Hour)
}
