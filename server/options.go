package server

import (
	logger "github.com/kuhufu/cm/logger"
	"time"
)

type optFunc func(s *Server)

func WithAuthTimeout(duration time.Duration) optFunc {
	return func(s *Server) {
		s.AuthTimeout = duration
	}
}

func WithHeartbeatTimeout(duration time.Duration) optFunc {
	return func(s *Server) {
		s.HeartbeatTimeout = duration
	}
}

func WithMessageHandler(handler MessageHandler) optFunc {
	return func(s *Server) {
		s.messageHandler = handler
	}
}

func WithDebugLog() optFunc {
	return func(s *Server) {
		logger.Init(logger.DebugLevel)
	}
}
