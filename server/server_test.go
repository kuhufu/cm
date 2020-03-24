package server

import (
	bytes2 "bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/kuhufu/cm/protocol"
	"github.com/kuhufu/cm/server/cm"
	"io/ioutil"
	"log"
	"testing"
	"time"
)

func Test(t *testing.T) {
	timer := time.NewTimer(time.Second)

	fmt.Println(timer.Reset(time.Second * 2))

	time.Sleep(time.Second * 3)
	fmt.Println(timer.Reset(time.Second * 2))
}

type Handle struct {
}

func (h Handle) Auth(data []byte) *AuthReply {

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

func (h Handle) PushIn(srcConn *cm.Conn, data []byte) (resp []byte) {
	return data
}

func (h Handle) OnConnClose(conn *cm.Conn) {

}

func TestServer_Run(t *testing.T) {
	srv := NewServer(
		WithMessageHandler(&Handle{}),
		WithAuthTimeout(time.Second*10),
		WithHeartbeatTimeout(time.Minute*100),
	)

	go func() {

		for {
			time.Sleep(time.Second)
			srv.PushToDeviceGroup("1", []byte("hello"))
		}
	}()

	err := srv.Run("ws", "0.0.0.0:8080")
	if err != nil {
		t.Error(err)
	}
}

func TestWsUpgrade(t *testing.T) {
	//development 123.56.103.77:7090
	//production kfws.qiyejiaoyou.com:7090
	f := func(uid string, os string) {
		conn, response, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
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

		msg := protocol.NewMessageWithDefault()
		msg.SetCmd(protocol.CmdAuth)
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
			log.Println(os + ":receive:", msg)
		}
	}

	go f("1", "web")
	go f("1", "android")


	time.Sleep(time.Hour)
}

func TestString(t *testing.T) {

}
