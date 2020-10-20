package server

type Handler interface {
	//不要修改data，方法返回后不能再使用data
	OnAuth(data []byte) *AuthReply
	OnReceive(channel *Channel, data []byte) (resp []byte)
	OnClose(channel *Channel)
}
