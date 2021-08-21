package msg

import (
	"bytes"
	"github.com/cryptopunkscc/astrald/components/sio"
	"io"
)

const StoryType = "message"

type Message struct {
	sender    []byte
	recipient []byte
	text      []byte
}

func (m *Message) Sender() string {
	return string(m.sender)
}

func (m *Message) Recipient() string {
	return string(m.recipient)
}

func (m *Message) Text() string {
	return string(m.text)
}

func NewMessage(
	sender string,
	recipient string,
	text string,
) *Message {
	return &Message{
		sender: bytes.NewBufferString(sender).Bytes(),
		recipient: bytes.NewBufferString(recipient).Bytes(),
		text: bytes.NewBufferString(text).Bytes(),
	}
}

func Unpack(b []byte) (m *Message, err error) {
	r := bytes.NewReader(b)
	return read(r)
}

func (m *Message) Pack() (r []byte) {
	b := bytes.NewBuffer(r)
	_ = m.write(b)
	return b.Bytes()
}

func read(r io.Reader) (m *Message, err error) {
	s := sio.NewReader(r)
	sender, err := s.ReadUint8()
	if err != nil {
		return
	}
	recipient, err := s.ReadUint8()
	if err != nil {
		return
	}
	text, err := s.ReadUint16()
	if err != nil {
		return
	}
	m = &Message{}

	if m.sender, err = s.ReadN(int(sender)); err != nil {
		return
	}
	if m.recipient, err = s.ReadN(int(recipient)); err != nil {
		return
	}
	if m.text, err = s.ReadN(int(text)); err != nil {
		return
	}

	return
}

func (m Message) write(w io.Writer) (err error) {
	s := sio.NewWriter(w)
	err = s.WriteUInt8(uint8(len(m.sender)))
	if err != nil {
		return
	}
	err = s.WriteUInt8(uint8(len(m.recipient)))
	if err != nil {
		return
	}
	_, err = s.WriteUInt16(uint16(len(m.text)))
	if err != nil {
		return
	}
	_, err = s.Write(m.sender)
	if err != nil {
		return
	}
	_, err = s.Write(m.recipient)
	if err != nil {
		return
	}
	_, err = s.Write(m.text)
	if err != nil {
		return
	}
	return
}