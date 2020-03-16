package listener

import (
	"github.com/kuhufu/cm/server/listener/ws"
	"net"
)

func Get(network, addr string) (net.Listener, error) {
	switch network {
	case "tcp", "tcp4", "tcp6":
		return net.Listen(network, addr)
	case "ws", "wss":
		return ws.Listen(network, addr)
	default:
		panic("不支持的协议类型")
	}
}
