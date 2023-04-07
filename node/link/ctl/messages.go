package ctl

import (
	"github.com/cryptopunkscc/astrald/cslq"
)

// message codes
const (
	mcClose = 0
	mcQuery = 1
	mcDrop  = 2
)

type Message interface{}

type CloseMessage struct {
}

type QueryMessage struct {
	query string
	port  int
}

func (m *QueryMessage) Query() string {
	return m.query
}

func (m *QueryMessage) Port() int {
	return m.port
}

func (m *QueryMessage) MarshalCSLQ(enc *cslq.Encoder) error {
	return enc.Encode("[c]c s", m.query, m.port)
}

func (m *QueryMessage) UnmarshalCSLQ(dec *cslq.Decoder) error {
	return dec.Decode("[c]c s", &m.query, &m.port)
}

type DropMessage struct {
	port int
}

func (m *DropMessage) Port() int {
	return m.port
}

func (m *DropMessage) MarshalCSLQ(enc *cslq.Encoder) error {
	return enc.Encode("c", m.port)
}

func (m *DropMessage) UnmarshalCSLQ(dec *cslq.Decoder) error {
	return dec.Decode("c", &m.port)
}
