package server

import (
	"crypto/tls"
	logger "github.com/kuhufu/cm/logger"
	"time"
)

type Options struct {
	CertFile  string
	KeyFile   string
	TlsConfig *tls.Config
}

type Option func(s *Server)

func WithAuthTimeout(duration time.Duration) Option {
	return func(s *Server) {
		s.AuthTimeout = duration
	}
}

func WithHeartbeatTimeout(duration time.Duration) Option {
	return func(s *Server) {
		s.HeartbeatTimeout = duration
	}
}

func WithMessageHandler(handler MessageHandler) Option {
	return func(s *Server) {
		s.messageHandler = handler
	}
}

func WithDebugLog() Option {
	return func(s *Server) {
		logger.Init(logger.DebugLevel)
	}
}

func WithCertAndKeyFile(cert, key string) Option {
	return func(s *Server) {
		s.opts.CertFile = cert
		s.opts.KeyFile = key
	}
}

func WithTlsConfig(config *tls.Config) Option {
	return func(s *Server) {
		s.opts.TlsConfig = config
	}
}
