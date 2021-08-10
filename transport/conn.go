package transport

import "net"

type BlockConn interface {
	net.Conn
	ReadBlock() ([]byte, error)
}
