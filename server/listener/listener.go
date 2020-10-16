package listener

import (
	"github.com/kuhufu/cm/server/listener/ws"
	"net"
	"net/url"
)

func Get(addr string) (net.Listener, error) {
	parse, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	scheme := parse.Scheme
	switch scheme {
	case "tcp", "tcp4", "tcp6":
		return net.Listen(scheme, parse.Host)
	case "ws", "wss":
		return ws.Listen(scheme, parse.Host+parse.Path)
	default:
		panic("不支持的协议类型")
	}
}
