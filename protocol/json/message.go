package json

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"github.com/kuhufu/cm/protocol/Interface"
	"github.com/kuhufu/cm/transport"
	"io"
	"io/ioutil"
)

const (
	KB = 1 << 10
	MB = KB << 10
)

const (
	DefaultMagicNumber = 0x08
	MsgLen             = 4
	MaxBodyLen         = 2 * MB
)

var (
	ErrBodyLenOverLimit = errors.New("body length over limit")
	ErrWrongBodyLen     = errors.New("wrong body length")
	ErrWrongMagicNumber = errors.New("wrong wrong magic number")
)

type MessageV1 struct {
	Message
}

func NewMessage() Interface.Message {
	return &MessageV1{}
}

func newMessage() *MessageV1 {
	return &MessageV1{}
}

func NewDefaultMessage() Interface.Message {
	return &MessageV1{
		Message: Message{
			MagicNumber: DefaultMagicNumber,
		},
	}
}

func (m *MessageV1) String() string {
	return m.Message.String()
}

func (m *MessageV1) ReadFrom(r io.Reader) (int64, error) {
	var data []byte
	var err error
	var n int

	switch r := r.(type) {
	case transport.BlockConn:
		data, err = r.ReadBlock()
	case *bytes.Reader:
		data, err = ioutil.ReadAll(r)
	default:
		msgLenBytes := make([]byte, MsgLen)
		n, err = r.Read(msgLenBytes)
		if n != MsgLen {
			return int64(n), ErrWrongBodyLen
		}

		if n > MaxBodyLen {
			return 0, ErrBodyLenOverLimit
		}

		bodyLen := binary.LittleEndian.Uint32(msgLenBytes)
		data = make([]byte, bodyLen)
		n, err = r.Read(data)
		if n != int(bodyLen) {
			return 0, ErrWrongBodyLen
		}
	}

	if err != nil {
		return 0, err
	}

	err = json.Unmarshal(data, &m.Message)

	if err := m.valid(); err != nil {
		return 0, err
	}

	return int64(len(data)), err
}

func (m *MessageV1) valid() error {
	return nil
}

func (m *MessageV1) WriteTo(w io.Writer) (int64, error) {
	var (
		err error
		n   int
	)

	marshal, err := json.Marshal(&m.Message)
	if err != nil {
		return 0, err
	}

	switch r := w.(type) {
	case transport.BlockConn:
		n, err = w.Write(marshal)
	default:
		msgLen := len(marshal)
		lenAndMsgBytes := make([]byte, MsgLen+msgLen)
		binary.LittleEndian.PutUint32(lenAndMsgBytes, uint32(msgLen))
		copy(lenAndMsgBytes[MsgLen:], marshal)

		n, err = r.Write(lenAndMsgBytes)
	}

	return int64(n), err
}

func (m *MessageV1) Decode(reader io.Reader) error {
	_, err := m.ReadFrom(reader)
	return err
}

func (m *MessageV1) Encode() []byte {
	marshal, _ := json.Marshal(&m.Message)
	return marshal
}

func (m *MessageV1) Size() uint32 {
	return 0
}

func (m *MessageV1) Cmd() Interface.Cmd {
	return Interface.Cmd(m.Message.Cmd)
}

func (m *MessageV1) Body() []byte {
	return []byte(m.Message.Body)
}

func (m *MessageV1) RequestId() uint32 {
	return m.Message.RequestId
}

func (m *MessageV1) SetRequestId(id uint32) Interface.Message {
	m.Message.RequestId = id
	return m
}

func (m *MessageV1) SetBody(body []byte) Interface.Message {
	m.Message.Body = string(body)
	return m
}

func (m *MessageV1) SetCmd(cmd Interface.Cmd) Interface.Message {
	m.Message.Cmd = Cmd(cmd)
	return m
}
