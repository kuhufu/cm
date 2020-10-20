package server

import "errors"

var (
	ErrRoomNotExist     = errors.New("room not exist")
	ErrAuthTimeout      = errors.New("auth timeout")      //认证超时
	ErrHeartbeatTimeout = errors.New("heartbeat timeout") //心跳超时
)
