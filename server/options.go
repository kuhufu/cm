package server

import (
	logger "github.com/kuhufu/cm/logger"
	"time"
)

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
