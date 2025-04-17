package nodes

import (
	"bytes"
	"encoding"
	"encoding/json"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

type StreamInfo struct {
	ID             astral.Int64
	LocalIdentity  *astral.Identity
	RemoteIdentity *astral.Identity
	LocalAddr      astral.String
	RemoteAddr     astral.String
	Outbound       astral.Bool
}

var _ astral.Object = &StreamInfo{}
var _ encoding.TextMarshaler = &StreamInfo{}
var _ json.Marshaler = &StreamInfo{}

func (s StreamInfo) ObjectType() string {
	return "mod.nodes.stream_info"
}

func (s StreamInfo) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(s).WriteTo(w)
}

func (s *StreamInfo) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(s).ReadFrom(r)
}

func (s StreamInfo) MarshalText() (text []byte, err error) {
	var b = &bytes.Buffer{}

	var d = "<"
	if s.Outbound {
		d = ">"
	}

	_, err = fmt.Fprintf(b, "%v %v %v %v", s.ID, d, s.RemoteIdentity, s.RemoteAddr)

	return b.Bytes(), err
}

// MarshalJSON is needed, because json marshaller will use MarshalText if this is absent
func (s StreamInfo) MarshalJSON() ([]byte, error) {
	type Alias StreamInfo
	a := Alias(s)
	return json.Marshal(a)
}
