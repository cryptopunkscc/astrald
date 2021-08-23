package uid

import (
	"bytes"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/sio"
	"io"
)

func Pack(c Card) []byte {
	b := bytes.NewBuffer([]byte{})
	_ = WriteCard(b, c)
	return b.Bytes()
}

func WriteCard(w io.Writer, c Card) (err error) {
	s := sio.NewWriter(w)
	if _, err = s.WriteStringWithSize8(string(c.Id)); err != nil {
		return
	}
	if _, err = s.WriteStringWithSize8(c.Alias); err != nil {
		return
	}
	if err = s.WriteUInt8(uint8(len(c.Endpoints))); err != nil {
		return
	}
	for _, e := range c.Endpoints {
		if err = WriteEndpoint(e, s); err != nil {
			return
		}
	}
	return
}

func WriteEndpoint(e Endpoint, w sio.Writer) (err error) {
	if _, err = w.WriteStringWithSize8(e.Network); err != nil {
		return
	}
	if _, err = w.WriteStringWithSize8(e.Address); err != nil {
		return
	}
	return
}

func Unpack(buff []byte) Card {
	b := bytes.NewBuffer(buff)
	c, _ := ReadCard(b)
	return c
}

func ReadCard(r io.Reader) (c Card, err error) {
	s := sio.NewReader(r)
	idString, err := s.ReadStringWithSize8()
	if err != nil {
		return
	}
	c.Id = api.Identity(idString)
	c.Alias, err = s.ReadStringWithSize8()
	if err != nil {
		return
	}
	size, err := s.ReadUint8()
	c.Endpoints, err = ReadEndpoints(s, size)
	return
}

func ReadEndpoints(r sio.Reader, n uint8) (es []Endpoint, err error) {
	es = make([]Endpoint, n)
	var e Endpoint
	for i := uint8(0); i < n; i++ {
		if e, err = ReadEndpoint(r); err != nil {
			return
		}
		es[i] = e
	}
	return
}

func ReadEndpoint(r sio.Reader) (e Endpoint, err error) {
	if e.Network, err = r.ReadStringWithSize8(); err != nil {
		return
	}
	if e.Address, err = r.ReadStringWithSize8(); err != nil {
		return
	}
	return
}
