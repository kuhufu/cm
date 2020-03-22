package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type Cmd uint32

const (
	CmdAuth = Cmd(iota + 1)
	CmdPush
	CmdHeartbeat
	CmdClose
	CmdServerPush
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
	MagicNumber = 0x08
	HeaderLen   = 20
	MaxBodyLen  = 2 * MB
)

//标记接口，需要一次性写入完整消息，否则头和body将分开写
type NeedFullWrite interface {
	MessageNeedFullWrite() bool
}

var (
	ErrBodyLenOverLimit = errors.New("body length over limit")
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
type Header [HeaderLen]byte

type Message struct {
	Header
	body []byte
}

func NewMessage() *Message {
	return &Message{
		body: nil,
	}
}

func newCustomMessage(cmd Cmd, body []byte) *Message {
	message := NewMessage()
	message.SetHeaderLen(HeaderLen)
	message.SetCmd(cmd)
	message.SetBody(body)
	return message
}

func NewMessageWithDefault() *Message {
	message := NewMessage()
	message.SetHeaderLen(HeaderLen)
	message.SetMagicNumber(MagicNumber)
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

func (m *Message) SetMagicNumber(n uint32) *Message {
	binary.BigEndian.PutUint32(m.Header[0:4], n)
	return m
}

func (m *Message) SetHeaderLen(n uint32) *Message {
	binary.BigEndian.PutUint32(m.Header[4:8], n)
	return m
}

func (m *Message) SetCmd(cmd Cmd) *Message {
	binary.BigEndian.PutUint32(m.Header[8:12], uint32(cmd))
	return m
}

func (m *Message) SetRequestId(n uint32) *Message {
	binary.BigEndian.PutUint32(m.Header[12:16], n)
	return m
}

func (m *Message) setBodyLen(n uint32) {
	binary.BigEndian.PutUint32(m.Header[16:20], n)
}

func (m *Message) MagicNumber() uint32 {
	return binary.BigEndian.Uint32(m.Header[0:4])
}

func (m *Message) HeaderLen() uint32 {
	return binary.BigEndian.Uint32(m.Header[4:8])
}

func (m *Message) Cmd() Cmd {
	return Cmd(binary.BigEndian.Uint32(m.Header[8:12]))
}

func (m *Message) RequestId() uint32 {
	return binary.BigEndian.Uint32(m.Header[12:16])
}

func (m *Message) BodyLen() uint32 {
	return binary.BigEndian.Uint32(m.Header[16:20])
}

func (m *Message) String() string {
	return fmt.Sprintf(
		`%v, "body":%s`,
		m.HeaderString(),
		m.Body(),
	)
}

func (m *Message) SetBody(data []byte) *Message {
	m.setBodyLen(uint32(len(data)))
	m.body = data
	return m
}

func (m *Message) Body() []byte {
	return m.body
}

func (m *Message) WriteTo(w io.Writer) error {
	if WriterNeedFullWrite(w) {
		data := m.Encode()
		_, err := w.Write(data)
		return err
	}

	_, err := w.Write(m.Header[:])
	if err != nil {
		return err
	}
	_, err = w.Write(m.body)
	return err
}

func (m *Message) Decode(r io.Reader) error {
	header := m.Header[:]

	//读取头部
	if _, err := io.ReadFull(r, header); err != nil {
		return err
	}

	if err := m.ValidHeader(); err != nil {
		return err
	}

	body := make([]byte, m.BodyLen())
	if _, err := io.ReadFull(r, body); err != nil {
		return err
	}
	m.SetBody(body)

	return nil
}

func (m *Message) ValidHeader() error {
	//检查magicNumber
	if m.MagicNumber() != MagicNumber {
		return ErrWrongMagicNumber
	}

	//检查header长度
	if m.HeaderLen() != HeaderLen {
		return ErrWrongHeaderLen
	}

	//限制body长度
	bodyLen := m.BodyLen()
	if bodyLen > MaxBodyLen {
		return ErrBodyLenOverLimit
	}

	return nil
}

func (m *Message) Encode() []byte {
	data := make([]byte, HeaderLen+m.BodyLen())

	copy(data[:HeaderLen], m.Header[:])
	copy(data[HeaderLen:], m.body)
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

func WriterNeedFullWrite(w io.Writer) bool {
	if v, ok := w.(NeedFullWrite); ok && v.MessageNeedFullWrite() {
		return true
	}
	return false
}
