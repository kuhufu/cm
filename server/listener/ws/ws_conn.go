package ws

import (
	"bytes"
	"errors"
	"github.com/gorilla/websocket"
	"io"
	"net"
	"sync"
	"time"
)

type Conn struct {
	conn   *websocket.Conn
	reader io.Reader

	rL sync.Mutex
	wL sync.Mutex
}

func (w *Conn) Read(b []byte) (n int, err error) {
	w.rL.Lock()
	defer w.rL.Unlock()

	reader := w.reader
	left := len(b)

	for left != 0 {
		if reader != nil {
			readn, err := reader.Read(b)
			if err != nil && err != io.EOF {
				return readn, err
			}
			left -= readn
		}

		if left != 0 {
			typ, data, err := w.conn.ReadMessage()
			if err != nil {
				return len(b) - left, err
			}
			if typ != websocket.BinaryMessage {
				return len(b) - left, errors.New("ws不是二进制消息")
			}

			reader = bytes.NewReader(data)
		}
	}

	w.reader = reader

	return len(b) - left, nil
}

func (w *Conn) Write(b []byte) (n int, err error) {
	w.wL.Lock()
	defer w.wL.Unlock()
	//writeMessage不是线程安全的
	err = w.conn.WriteMessage(websocket.BinaryMessage, b)
	if err != nil {
		return 0, err
	}
	return len(b), nil
}

func (w *Conn) Close() error {
	return w.conn.Close()
}

func (w *Conn) LocalAddr() net.Addr {
	return w.conn.LocalAddr()
}

func (w *Conn) RemoteAddr() net.Addr {
	return w.conn.RemoteAddr()
}

func (w *Conn) SetDeadline(t time.Time) error {
	err := w.conn.SetReadDeadline(t)
	if err != nil {
		return err
	}
	err = w.conn.SetWriteDeadline(t)
	return err
}

func (w *Conn) SetReadDeadline(t time.Time) error {
	return w.conn.SetReadDeadline(t)
}

func (w *Conn) SetWriteDeadline(t time.Time) error {
	return w.conn.SetWriteDeadline(t)
}

//protocol的标记接口，标记是否需要一次性写入整个消息
func (w *Conn) MessageNeedFullWrite() bool {
	return true
}
