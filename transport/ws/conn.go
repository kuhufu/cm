package ws

import (
	"bytes"
	"errors"
	"github.com/gorilla/websocket"
	"io"
	"sync"
	"time"
)

type Conn struct {
	*websocket.Conn
	reader io.Reader

	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	rL sync.Mutex
	wL sync.Mutex
}

func (c *Conn) Read(b []byte) (n int, err error) {
	if c.ReadTimeout != 0 {
		err = c.SetReadDeadline(time.Now().Add(c.ReadTimeout))
		if err != nil {
			return 0, err
		}
	}

	c.rL.Lock()
	defer c.rL.Unlock()

	reader := c.reader
	left := len(b)

	for left != 0 {
		if reader != nil {
			n, err := reader.Read(b)
			if err != nil && err != io.EOF {
				return n, err
			}
			left -= n
		}

		if left != 0 {
			typ, data, err := c.ReadMessage()
			if err != nil {
				return len(b) - left, err
			}
			if typ != websocket.BinaryMessage {
				return len(b) - left, errors.New("ws不是二进制消息")
			}

			reader = bytes.NewReader(data)
		}
	}

	c.reader = reader

	return len(b) - left, nil
}

func (c *Conn) Write(b []byte) (n int, err error) {
	if c.WriteTimeout != 0 {
		err = c.SetWriteDeadline(time.Now().Add(c.ReadTimeout))
		if err != nil {
			return 0, err
		}
	}

	c.wL.Lock()
	defer c.wL.Unlock()
	//writeMessage不是线程安全的
	err = c.WriteMessage(websocket.BinaryMessage, b)
	if err != nil {
		return 0, err
	}
	return len(b), nil
}

func (c *Conn) SetDeadline(t time.Time) error {
	err := c.SetReadDeadline(t)
	if err != nil {
		return err
	}
	err = c.SetWriteDeadline(t)
	return err
}

//protocol的标记接口，标记是否需要一次性写入整个消息
func (c *Conn) MessageNeedFullWrite() bool {
	return true
}
