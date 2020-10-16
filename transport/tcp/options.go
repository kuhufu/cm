package tcp

import (
	"crypto/tls"
)

type Options struct {
	CertFile  string
	KeyFile   string
	TlsConfig *tls.Config
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
