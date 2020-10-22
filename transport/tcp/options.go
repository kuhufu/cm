package tcp

import (
	"crypto/tls"
	"time"
)

//TlsConfig 和 证书文件文件路径 二选一， 同时设置，将选择TlsConfig
type Options struct {
	CertFile     string
	KeyFile      string
	TlsConfig    *tls.Config
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func (opts *Options) Init() error {
	config := opts.TlsConfig
	certFile := opts.CertFile
	keyFile := opts.KeyFile

	configHasCert := config != nil && (len(config.Certificates) > 0 || config.GetCertificate != nil)
	if !configHasCert && (certFile != "" || keyFile != "") {
		if config == nil {
			config = &tls.Config{}
		}
		var err error
		config.Certificates = make([]tls.Certificate, 1)
		config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return err
		}
	}
	opts.TlsConfig = config

	return nil
}

type Option func(options *Options)

func WithTLsConfig(config *tls.Config) Option {
	return func(options *Options) {
		options.TlsConfig = config
	}
}
