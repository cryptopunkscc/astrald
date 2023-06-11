package ctl

import (
	"github.com/cryptopunkscc/astrald/cslq"
)

// message codes
const (
	mcClose = 0
	mcQuery = 1
	mcDrop  = 2
	mcPing  = 3
)

type Message interface{}

type CloseMessage struct {
}

type PingMessage struct {
	port int
}

func (m *PingMessage) Port() int {
	return m.port
}

func (m *PingMessage) MarshalCSLQ(enc *cslq.Encoder) error {
	return enc.Encodef("s", m.port)
}

func (m *PingMessage) UnmarshalCSLQ(dec *cslq.Decoder) error {
	return dec.Decodef("s", &m.port)
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
	return enc.Encodef("[c]c s", m.query, m.port)
}

func (m *QueryMessage) UnmarshalCSLQ(dec *cslq.Decoder) error {
	return dec.Decodef("[c]c s", &m.query, &m.port)
}

type DropMessage struct {
	port int
}

func (m *DropMessage) Port() int {
	return m.port
}

func (m *DropMessage) MarshalCSLQ(enc *cslq.Encoder) error {
	return enc.Encodef("c", m.port)
}

func (m *DropMessage) UnmarshalCSLQ(dec *cslq.Decoder) error {
	return dec.Decodef("c", &m.port)
}
