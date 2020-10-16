package protobuf

import (
	"encoding/binary"
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/kuhufu/cm/protocol/Interface"
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
	size [4]byte
	msg  Message
}

func NewMessage() *MessageV1 {
	return &MessageV1{}
}

func NewDefaultMessage() *MessageV1 {
	return &MessageV1{
		msg: Message{
			MagicNumber: DefaultMagicNumber,
		},
	}
}

func (m *MessageV1) String() string {
	return m.msg.String()
}

func (m *MessageV1) ReadFrom(r io.Reader) (int64, error) {
	buf := m.size[:]
	n, err := r.Read(m.size[:])
	if err != nil {
		return int64(n), err
	}

	binary.LittleEndian.Uint32(m.size[:])

	if err := m.valid(); err != nil {
		return 0, err
	}

	buf = make([]byte, m.Size())
	n, err = r.Read(buf)
	if err != nil {
		return int64(n), err
	}

	err = proto.Unmarshal(buf, &m.msg)

	return int64(n), err
}

func (m *MessageV1) valid() error {
	if m.Size() > MaxBodyLen+20 {
		return ErrBodyLenOverLimit
	}

	if m.Size() < 0 {
		return ErrWrongBodyLen
	}

	if m.msg.MagicNumber != DefaultMagicNumber {
		return ErrWrongMagicNumber
	}

	return nil
}

func (m *MessageV1) WriteTo(w io.Writer) (int64, error) {
	marshal, err := proto.Marshal(&m.msg)
	if err != nil {
		return 0, err
	}

	n, err := w.Write(marshal)
	return int64(n), err
}

func (m *MessageV1) Decode(reader io.Reader) error {
	_, err := m.ReadFrom(reader)
	return err
}

func (m *MessageV1) Encode() []byte {
	marshal, _ := proto.Marshal(&m.msg)
	return marshal
}

func (m *MessageV1) Size() uint32 {
	return binary.LittleEndian.Uint32(m.size[:])
}

func (m *MessageV1) Cmd() uint32 {
	return uint32(m.msg.Cmd)
}

func (m *MessageV1) Body() []byte {
	return m.msg.Body
}

func (m *MessageV1) RequestId() uint32 {
	return m.msg.RequestId
}

func (m *MessageV1) SetRequestId(id uint32) Interface.Message {
	m.msg.RequestId = id
	return m
}

func (m *MessageV1) SetBody(body []byte) Interface.Message {
	m.msg.Body = body
	return m
}

func (m *MessageV1) SetCmd(cmd uint32) Interface.Message {
	m.msg.Cmd = Cmd(cmd)
	return m
}
