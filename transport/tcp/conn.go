package tcp

import (
	"net"
	"time"
)

type Conn struct {
	net.Conn
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func (c *Conn) Read(b []byte) (n int, err error) {
	if c.ReadTimeout != 0 {
		err = c.SetReadDeadline(time.Now().Add(c.ReadTimeout))
		if err != nil {
			return 0, err
		}
	}

	return c.Conn.Read(b)
}

func (c *Conn) Write(b []byte) (n int, err error) {
	if c.WriteTimeout != 0 {
		err = c.SetWriteDeadline(time.Now().Add(c.WriteTimeout))
		if err != nil {
			return 0, err
		}
	}

	return c.Conn.Write(b)
}
