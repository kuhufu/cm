package server

type AuthReply struct {
	Ok        bool
	RoomId    string
	ChannelId string //不能为空，否则panic
	Data      []byte
	Metadata  map[interface{}]interface{}
	err       error
}
