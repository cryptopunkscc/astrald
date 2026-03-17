package nodes

import (
	"bytes"
	"encoding"
	"encoding/json"
	"fmt"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

type StreamInfo struct {
	ID              astral.Nonce
	LocalIdentity   *astral.Identity
	RemoteIdentity  *astral.Identity
	LocalEndpoint   exonet.Endpoint
	RemoteEndpoint  exonet.Endpoint
	Outbound        astral.Bool
	Network         astral.String8
	HighPressure    astral.Bool
	BytesThroughput astral.Uint64
}

var _ astral.Object = &StreamInfo{}
var _ encoding.TextMarshaler = &StreamInfo{}
var _ json.Marshaler = &StreamInfo{}
var _ json.Unmarshaler = &StreamInfo{}

func (s StreamInfo) ObjectType() string {
	return "mod.nodes.stream_info"
}

func (s StreamInfo) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Objectify(&s).WriteTo(w)
}

func (s *StreamInfo) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Objectify(s).ReadFrom(r)
}

func (s StreamInfo) MarshalText() (text []byte, err error) {
	var b = &bytes.Buffer{}

	var d = "<"
	if s.Outbound {
		d = ">"
	}

	_, err = fmt.Fprintf(b, "%v %v %v %v throughput=%v high=%v", s.ID, d, s.RemoteIdentity, s.RemoteEndpoint, s.BytesThroughput, bool(s.HighPressure))

	return b.Bytes(), err
}

func (e StreamInfo) MarshalJSON() ([]byte, error) {
	return astral.Objectify(&e).MarshalJSON()
}

func (e *StreamInfo) UnmarshalJSON(b []byte) error {
	return astral.Objectify(e).UnmarshalJSON(b)
}

func init() {
	astral.Add(&StreamInfo{})
}
