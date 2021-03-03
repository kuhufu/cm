package test

import (
	bytes2 "bytes"
	"crypto/tls"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/kuhufu/cm/protocol"
	"github.com/kuhufu/cm/protocol/consts"
	"io/ioutil"
	"log"
	"testing"
	"time"
)

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

			_, err = msg.ReadFrom(bytes2.NewReader(data))
			if err != nil {
				log.Println(err)
				return
			}
			log.Println(os+":receive:", msg)
		}
	}

	go f("1", "web")

	time.Sleep(time.Hour)
}
