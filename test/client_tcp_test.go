package test

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/kuhufu/cm/protocol"
	"github.com/kuhufu/cm/protocol/consts"
	"log"
	"os"
	"testing"
	"time"
)

func Test_ClientTcp(t *testing.T) {
	pemData, err := os.ReadFile("cert/cert.pem")
	if err != nil {
		panic(err)
	}
	block, _ := pem.Decode(pemData)
	certificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Error(err)
		return
	}

	ErrBadCertificate := errors.New("bad certificate")

	conf := &tls.Config{
		InsecureSkipVerify: true,

		VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			t.Log("VerifyPeerCertificate")
			ok := bytes.Equal(certificate.Raw, rawCerts[0])
			if !ok {
				return ErrBadCertificate
			}

			return nil
		},
	}

	//conf := &tls.Config{
	//	InsecureSkipVerify: true,
	//}

	factory := protocol.GetFactory(protocol.BINARY)

	f := func(uid string, os string) {
		conn, err := tls.Dial("tcp", "localhost:8080", conf)
		if err != nil {
			t.Error(err)
			return
		}

		fmt.Println(conn.RemoteAddr())

		time.Sleep(time.Millisecond)

		msg := factory.NewDefaultMessage()
		msg.SetCmd(consts.CmdAuth)
		msg.SetBody([]byte(fmt.Sprintf(`{"uid":"%v","os":"%v"}`, uid, os)))

		fmt.Println(msg)

		_, err = msg.WriteTo(conn)
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
