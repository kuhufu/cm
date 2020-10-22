package server

import (
	"crypto/tls"
	logger "github.com/kuhufu/cm/logger"
	"time"
)

type Options struct {
	//认证超时时间，默认10s
	AuthTimeout time.Duration
	//心跳超时时间，默认90s
	HeartbeatTimeout time.Duration
	//证书文件路径
	CertFile string
	//key文件路径
	KeyFile string
	//tls配置
	TlsConfig *tls.Config
	//handler
	Handler Handler
	//读超时时间，0表示不超时
	ReadTimeout time.Duration
	//写超时时间，0表示不超时
	WriteTimeout time.Duration
}

func defaultOptions() Options {
	return Options{
		AuthTimeout:      time.Second * 10,
		HeartbeatTimeout: time.Second * 90,
	}
}

type Option func(o *Options)

func WithReadTimeout(duration time.Duration) Option {
	return func(o *Options) {
		o.ReadTimeout = duration
	}
}

func WithWriteTimeout(duration time.Duration) Option {
	return func(o *Options) {
		o.WriteTimeout = duration
	}
}

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
