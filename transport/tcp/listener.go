package tcp

import (
	"crypto/tls"
	"net"
)

type Listener struct {
	net.Listener
}

func Listen(network, addr string, opts Options) (net.Listener, error) {
	if err := opts.Init(); err != nil {
		return nil, err
	}

	ln, err := net.Listen(network, addr)
	if err != nil {
		return nil, err
	}

	if opts.TlsConfig != nil {
		ln = tls.NewListener(ln, opts.TlsConfig)
	}
	return ln, nil
}
