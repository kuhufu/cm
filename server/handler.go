package server

type Handler interface {
	OnAuth(data []byte) *AuthReply
	OnReceive(channel *Channel, data []byte) (resp []byte)
	OnClose(channel *Channel)
}
