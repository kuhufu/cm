package test

import (
	"crypto/tls"
	"fmt"
	"github.com/kuhufu/cm/protocol"
	"github.com/kuhufu/cm/protocol/consts"
	"log"
	"testing"
	"time"
)

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
			_, err := msg.ReadFrom(conn)
			if err != nil {
				log.Println(err)
				return
			}
			log.Println(os+":receive:", msg)
		}
	}

	go f("2", "web")

	time.Sleep(time.Hour)
}
