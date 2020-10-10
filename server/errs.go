package server

import "errors"

var (
	ErrConnNotExist     = errors.New("connect not exist")
	ErrAuthTimeout      = errors.New("auth timeout")      //认证超时
	ErrHeartbeatTimeout = errors.New("heartbeat timeout") //心跳超时
)
