package server

import (
	"crypto/tls"
	logger "github.com/kuhufu/cm/logger"
	"time"
)

type Options struct {
	//认证超时时间
	AuthTimeout time.Duration
	//心跳超时时间
	HeartbeatTimeout time.Duration
	//证书文件路径
	CertFile string
	//key文件路径
	KeyFile string
	//tls配置
	TlsConfig *tls.Config
	//handler
	Handler Handler
}

type Option func(o *Options)

func WithAuthTimeout(duration time.Duration) Option {
	return func(o *Options) {
		o.AuthTimeout = duration
	}
}

func WithHeartbeatTimeout(duration time.Duration) Option {
	return func(o *Options) {
		o.HeartbeatTimeout = duration
	}
}

func WithHandler(handler Handler) Option {
	return func(o *Options) {
		o.Handler = handler
	}
}

func WithDebugLog() Option {
	return func(o *Options) {
		logger.Init(logger.DebugLevel)
	}
}

func WithCertAndKeyFile(cert, key string) Option {
	return func(o *Options) {
		o.CertFile = cert
		o.KeyFile = key
	}
}

func WithTlsConfig(config *tls.Config) Option {
	return func(o *Options) {
		o.TlsConfig = config
	}
}
