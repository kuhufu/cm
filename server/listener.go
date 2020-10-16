package server

import (
	"github.com/kuhufu/cm/transport/tcp"
	"github.com/kuhufu/cm/transport/ws"
	"net"
	"net/url"
)

func GetListener(addr string, options Options) (net.Listener, error) {
	parse, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	scheme := parse.Scheme
	switch scheme {
	case "tcp", "tcp4", "tcp6":
		opts := tcp.Options{
			CertFile:  options.CertFile,
			KeyFile:   options.KeyFile,
			TlsConfig: options.TlsConfig,
		}

		return tcp.Listen(scheme, parse.Host, opts)
	case "ws", "wss":
		opts := ws.Options{
			CertFile:  options.CertFile,
			KeyFile:   options.KeyFile,
			TlsConfig: options.TlsConfig,
		}
		return ws.Listen(scheme, parse.Host+parse.Path, opts)
	default:
		panic("不支持的协议类型")
	}
}
