package ws

import (
	"fmt"
	"github.com/kuhufu/cm/protocol"
	"io"
	"testing"
)

func TestServer_HandleWriter(t *testing.T) {
	var m map[string]string

	if _, ok := m["sdf"]; ok {
		fmt.Println("hhh")
	}
}

func TestTypeCheck(t *testing.T) {
	conn := &Conn{}
	w := io.Writer(conn)

	if _, ok := w.(protocol.NeedFullWrite); !ok {
		t.Error("fail")
	}
}
