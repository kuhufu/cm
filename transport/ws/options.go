package ws

import (
	"crypto/tls"
	"net/http"
)

//TlsConfig 和 证书文件文件路径 二选一， 同时设置，将选择TlsConfig
type Options struct {
	CertFile  string
	KeyFile   string
	TlsConfig *tls.Config
	ServeMux  *http.ServeMux
}

func (opts *Options) Init() error {
	if opts.ServeMux == nil {
		opts.ServeMux = http.DefaultServeMux
	}

	return nil
}

type Option func(options *Options)

func WithCertAndKey(cert, key string) Option {
	return func(options *Options) {
		options.CertFile = cert
		options.KeyFile = key
	}
}

func WithServeMux(mux *http.ServeMux) Option {
	return func(options *Options) {
		options.ServeMux = mux
	}
}
