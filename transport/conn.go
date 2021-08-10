package transport

import "net"

type BlockConn interface {
	net.Conn
	ReadBlock() ([]byte, error)
	WriteString(str []byte) error
}
