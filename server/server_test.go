package server

import (
	bytes2 "bytes"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/kuhufu/cm/protocol"
	"io/ioutil"
	"log"
	"sync"
	"testing"
	"time"
	"unsafe"
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
	return &AuthReply{
		Ok:       true,
		ConnId:   "22",
		Data:     []byte("hello"),
		AuthTime: time.Now(),
	}
}

func (h Handle) PushIn(srcConn *Conn, data []byte) (resp []byte) {
	return data
}

func (h Handle) OnConnClose(conn *Conn) {

}

func TestServer_Run(t *testing.T) {
	srv := NewServer(
		WithMessageHandler(&Handle{}),
		WithAuthTimeout(time.Second*10),
		WithHeartbeatTimeout(time.Minute),
	)

	err := srv.Run("ws", "0.0.0.0:8080")
	if err != nil {
		t.Error(err)
	}
}

func TestWsUpgrade(t *testing.T) {
	//development 123.56.103.77:7090
	//production kfws.qiyejiaoyou.com:7090
	f := func() {
		conn, response, err := websocket.DefaultDialer.Dial("ws://123.56.103.77:7090/ws", nil)
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
		msg.SetBody([]byte(`{"token":"4fac133ddc96863d8fcd8f725c04ba2ed9e6ef486b5287afcbd1e15a0e9e8563"}`))

		fmt.Println(msg)

		err = conn.WriteMessage(websocket.BinaryMessage, msg.Encode())
		if err != nil {
			t.Error(err)
			return
		}

		messageType, data, err := conn.ReadMessage()
		if err != nil {
			log.Println()
			t.Error(err)
			return
		}

		if messageType != websocket.BinaryMessage {
			t.Errorf("wrong messageType: %v", messageType)
			return
		}

		err = msg.Decode(bytes2.NewReader(data))
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Println("receive data:", msg)

		for i := 0; i < 2; i++ {
			//写

			time.Sleep(time.Second)

			msg := protocol.NewMessageWithDefault()
			msg.SetCmd(protocol.CmdPush)
			msg.SetBody([]byte("push"))
			err := conn.WriteMessage(websocket.BinaryMessage, msg.Encode())
			if err != nil {
				t.Error(err)
				return
			}
			log.Println("write:", msg)
			//读
			messageType, data, err := conn.ReadMessage()
			if err != nil {
				log.Println()
				t.Error(err)
				return
			}

			if messageType != websocket.BinaryMessage {
				t.Errorf("wrong messageType: %v", messageType)
				return
			}

			msg.Decode(bytes2.NewReader(data))
			log.Println("receive:", msg)
		}
	}

	wg := sync.WaitGroup{}

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			f()
			wg.Done()
		}()
	}

	wg.Wait()
}

func TestString(t *testing.T) {
	t1 := time.Now()
	t2 := time.Now()

	fmt.Println(uintptr(unsafe.Pointer(&t1)))
	fmt.Println(uintptr(unsafe.Pointer(&t2)))

}
