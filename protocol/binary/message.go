package binary

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/kuhufu/cm/protocol/Interface"
	"io"
)

type Cmd uint32

const (
	CmdUnknown    = Cmd(0)
	CmdAuth       = Cmd(1)
	CmdPush       = Cmd(2)
	CmdHeartbeat  = Cmd(3)
	CmdClose      = Cmd(4)
	CmdServerPush = Cmd(5)
)

const (
	KB = 1 << 10
	MB = KB << 10
)

var cmdMap = map[Cmd]string{
	CmdAuth:       "CmdAuth",
	CmdPush:       "CmdPush",
	CmdHeartbeat:  "CmdHeartbeat",
	CmdClose:      "CmdClose",
	CmdServerPush: "CmdServerPush",
}

const (
	DefaultMagicNumber = 0x08
	DefaultHeaderLen   = 20
	MaxBodyLen         = 2 * MB
)

var (
	ErrBodyLenOverLimit = errors.New("body length over limit")
	ErrWrongBodyLen     = errors.New("wrong body length")
	ErrWrongHeaderLen   = errors.New("wrong header length")
	ErrWrongMagicNumber = errors.New("wrong wrong magic number")
)

func (c Cmd) String() string {
	return cmdMap[c]
}

//magicNumber uint32
//headerLen   uint32
//cmd         Cmd
//requestId   uint32 请求id由客户端设置
//bodyLen     uint32
type header [DefaultHeaderLen]byte

type Message struct {
	header
	body []byte
}

func NewMessage() *Message {
	return &Message{
		body: nil,
	}
}

func newCustomMessage(cmd Cmd, body []byte) *Message {
	message := NewMessage()
	message.SetHeaderLen(DefaultHeaderLen)
	message.SetCmd(uint32(cmd))
	message.SetBody(body)
	return message
}

func NewDefaultMessage() *Message {
	message := NewMessage()
	message.SetHeaderLen(DefaultHeaderLen)
	message.SetMagicNumber(DefaultMagicNumber)
	return message
}

func (m *Message) HeaderString() string {
	return fmt.Sprintf(
		`"magicNumber":%v, "headerLen":%v, "cmd":%v, "requestId":%v, bodyLen":%v`,
		m.MagicNumber(),
		m.HeaderLen(),
		m.Cmd(),
		m.RequestId(),
		m.BodyLen(),
	)
}

func (m *Message) SetMagicNumber(n uint32) Interface.Message {
	binary.BigEndian.PutUint32(m.header[0:4], n)
	return m
}

func (m *Message) SetHeaderLen(n uint32) Interface.Message {
	binary.BigEndian.PutUint32(m.header[4:8], n)
	return m
}

func (m *Message) SetCmd(cmd uint32) Interface.Message {
	binary.BigEndian.PutUint32(m.header[8:12], cmd)
	return m
}

func (m *Message) SetRequestId(n uint32) Interface.Message {
	binary.BigEndian.PutUint32(m.header[12:16], n)
	return m
}

func (m *Message) setBodyLen(n uint32) {
	binary.BigEndian.PutUint32(m.header[16:20], n)
}

func (m *Message) MagicNumber() uint32 {
	return binary.BigEndian.Uint32(m.header[0:4])
}

func (m *Message) HeaderLen() uint32 {
	return binary.BigEndian.Uint32(m.header[4:8])
}

func (m *Message) Cmd() uint32 {
	return binary.BigEndian.Uint32(m.header[8:12])
}

func (m *Message) RequestId() uint32 {
	return binary.BigEndian.Uint32(m.header[12:16])
}

func (m *Message) BodyLen() uint32 {
	return binary.BigEndian.Uint32(m.header[16:20])
}

func (m *Message) String() string {
	return fmt.Sprintf(
		`%v, "body":%s`,
		m.HeaderString(),
		m.Body(),
	)
}

func (m *Message) SetBody(data []byte) Interface.Message {
	m.setBodyLen(uint32(len(data)))
	m.body = data
	return m
}

func (m *Message) Body() []byte {
	return m.body
}

func (m *Message) ReadFrom(r io.Reader) (int64, error) {
	header := m.header[:]

	//读取头部
	n, err := io.ReadFull(r, header)
	if err != nil {
		return int64(n), err
	}

	if err := m.validHeader(); err != nil {
		return int64(n), err
	}

	//一个小优化
	body := m.body
	bodyLen := int(m.BodyLen())
	if cap(body) < bodyLen {
		body = make([]byte, bodyLen)
	} else {
		body = body[:bodyLen]
	}

	n, err = io.ReadFull(r, body)
	if err != nil {
		return int64(n), err
	}
	m.SetBody(body)

	return 0, nil
}

func (m *Message) WriteTo(w io.Writer) (int64, error) {
	if writerNeedFullWrite(w) {
		data := m.Encode()
		n, err := w.Write(data)
		return int64(n), err
	}

	n, err := w.Write(m.header[:])
	if err != nil {
		return int64(n), err
	}
	nb, err := w.Write(m.body)
	return int64(n + nb), err
}

func (m *Message) validHeader() error {
	//检查magicNumber
	if m.MagicNumber() != DefaultMagicNumber {
		return ErrWrongMagicNumber
	}

	//检查header长度
	if m.HeaderLen() != DefaultHeaderLen {
		return ErrWrongHeaderLen
	}

	//限制body长度
	bodyLen := m.BodyLen()
	if bodyLen > MaxBodyLen {
		return ErrBodyLenOverLimit
	}

	if bodyLen < 0 {
		return ErrWrongBodyLen
	}

	return nil
}

func (m *Message) Decode(r io.Reader) error {
	_, err := m.ReadFrom(r)
	return err
}

func (m *Message) Encode() []byte {
	data := make([]byte, DefaultHeaderLen+m.BodyLen())

	copy(data[:DefaultHeaderLen], m.header[:])
	copy(data[DefaultHeaderLen:], m.body)
	return data
}

func Read(r io.Reader) (*Message, error) {
	message := NewMessage()

	err := message.Decode(r)
	if err != nil {
		return nil, err
	}

	return message, nil
}

func writerNeedFullWrite(w io.Writer) bool {
	if v, ok := w.(Interface.NeedFullWrite); ok && v.MessageNeedFullWrite() {
		return true
	}
	return false
}
