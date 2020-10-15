package protobuf

import (
	"encoding/binary"
	"github.com/golang/protobuf/proto"
	"github.com/kuhufu/cm/protocol"
	"io"
)

type MessageV1 struct {
	size [4]byte
	msg  Message
}

func NewMessage() *MessageV1 {
	return &MessageV1{}
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
	if m.Size() > protocol.MaxBodyLen+20 {
		return protocol.ErrBodyLenOverLimit
	}

	if m.Size() < 0 {
		return protocol.ErrWrongBodyLen
	}

	if m.msg.MagicNumber != protocol.DefaultMagicNumber {
		return protocol.ErrWrongMagicNumber
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

func (m *MessageV1) Encode(writer io.Writer) error {
	_, err := m.WriteTo(writer)
	return err
}

func (m *MessageV1) Size() uint32 {
	return binary.LittleEndian.Uint32(m.size[:])
}

func (m *MessageV1) Cmd() Cmd {
	return m.msg.Cmd
}

func (m *MessageV1) Body() []byte {
	return m.msg.Body
}
