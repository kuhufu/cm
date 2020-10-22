package tcp

import (
	"crypto/tls"
	"net"
)

type Listener struct {
	opts Options
	net.Listener
}

func (l *Listener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	return &Conn{
		Conn:         conn,
		ReadTimeout:  l.opts.ReadTimeout,
		WriteTimeout: l.opts.WriteTimeout,
	}, nil
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

	return &Listener{
		Listener: ln,
		opts:     opts,
	}, nil
}
