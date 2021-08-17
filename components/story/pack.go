package story

import (
	"bytes"
	"github.com/cryptopunkscc/astrald/components/serialize"
	"io"
)

func (s *Story) PackBytes() []byte {
	buffer := bytes.NewBuffer([]byte{})
	_ = s.Pack(buffer)
	return buffer.Bytes()
}


func (s *Story) Pack(writer io.Writer) (err error) {
	f := serialize.NewFormatter(writer)

	// write magic bytes
	_, err = f.Write(MagicBytes[:])
	if err != nil {
		return
	}

	// write timestamp
	_, err = f.WriteUInt64(s.timestamp)
	if err != nil {
		return
	}

	// write sizes
	err = f.WriteByte(byte(len(s.typ)))
	if err != nil {
		return
	}
	err = f.WriteByte(byte(len(s.author)))
	if err != nil {
		return
	}
	err = f.WriteByte(byte(len(s.Refs())))
	if err != nil {
		return
	}
	_, err = f.WriteUInt16(uint16(len(s.Data())))
	if err != nil {
		return
	}

	// write values
	_, err = f.Write(s.typ)
	if err != nil {
		return
	}
	_, err = f.Write(s.author)
	if err != nil {
		return
	}
	for _, ref := range s.refs {
		id := ref.Pack()
		_, err = f.Write(id[:])
		if err != nil {
			return
		}
	}
	if len(s.data) == 0 {
		return
	}
	_, err = f.Write(s.data)
	if err != nil {
		return
	}

	return
}
