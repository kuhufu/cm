package binary

import (
	"bytes"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"
	"unsafe"
)

func TestMessage_Encode(t *testing.T) {
	m := NewDefaultMessage().SetCmd(CmdAuth).SetBody([]byte("111"))
	encode := m.Encode()

	m2, err := Read(bytes.NewReader(encode))
	if err != nil {
		t.Error(err)
	}

	bodyEqual := string(m.body) == string(m2.body)
	headerEqual := m.header == m2.header

	if !bodyEqual || !headerEqual {
		t.Error("消息编码错误")
	}
}

func TestMessage_Decode(t *testing.T) {

}

func TestMessage_print(t *testing.T) {
	fmt.Println(KB)
	fmt.Println(MB)
}

func TestPrintMessage(t *testing.T) {
	fmt.Printf("%+v", newCustomMessage(CmdAuth, []byte("body")))
}

type BigData [1024 * 1024 * 128]byte

var p = sync.Pool{
	New: func() interface{} {
		return &BigData{}
	},
}

func TestGetPoolMsg(t *testing.T) {
	arr := []interface{}{}
	for i := 0; i < 4; i++ {
		arr = append(arr, p.Get())
	}
	fmt.Println(unsafe.Pointer(&arr))
	fmt.Println("获取完成")

	fmt.Println("第一次gc")
	runtime.GC()

	time.Sleep(time.Second * 10)

	fmt.Println(unsafe.Pointer(&arr))
	//arr = nil
	fmt.Println("第二次GC")
	runtime.GC()

	time.Sleep(time.Second * 3000)

	fmt.Println(&arr)
}
