package json

import (
	"encoding/json"
	"errors"
	"github.com/kuhufu/cm/protocol/Interface"
	"github.com/kuhufu/cm/transport"
	"io"
)

const (
	KB = 1 << 10
	MB = KB << 10
)

const (
	DefaultMagicNumber = 0x08
	DefaultHeaderLen   = 20
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

func NewMessage() *MessageV1 {
	return &MessageV1{}
}

func NewDefaultMessage() *MessageV1 {
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

	switch r := r.(type) {
	case transport.BlockConn:
		data, err = r.ReadBlock()
	default:
		data, err = io.ReadAll(r)
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

	n, err = w.Write(marshal)

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
