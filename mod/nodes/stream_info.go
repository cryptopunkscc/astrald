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

func marshalEndpointJSON(e exonet.Endpoint) (astral.JSONAdapter, error) {
	if e == nil {
		return astral.JSONAdapter{}, nil
	}
	b, err := json.Marshal(e)
	if err != nil {
		return astral.JSONAdapter{}, err
	}
	return astral.JSONAdapter{Type: e.ObjectType(), Object: b}, nil
}

func unmarshalEndpointJSON(a astral.JSONAdapter) (exonet.Endpoint, error) {
	if a.Type == "" {
		return nil, nil
	}
	obj := astral.New(a.Type)
	if obj == nil {
		return nil, astral.NewErrBlueprintNotFound(a.Type)
	}
	if err := json.Unmarshal(a.Object, obj); err != nil {
		return nil, err
	}
	ep, ok := obj.(exonet.Endpoint)
	if !ok {
		return nil, fmt.Errorf("StreamInfo: %v is not an exonet.Endpoint", a.Type)
	}
	return ep, nil
}

type streamInfoJSON struct {
	ID              astral.Nonce
	LocalIdentity   *astral.Identity
	RemoteIdentity  *astral.Identity
	LocalEndpoint   astral.JSONAdapter
	RemoteEndpoint  astral.JSONAdapter
	Outbound        astral.Bool
	Network         astral.String8
	HighPressure    astral.Bool
	BytesThroughput astral.Uint64
}

func (s StreamInfo) MarshalJSON() ([]byte, error) {
	localEP, err := marshalEndpointJSON(s.LocalEndpoint)
	if err != nil {
		return nil, err
	}
	remoteEP, err := marshalEndpointJSON(s.RemoteEndpoint)
	if err != nil {
		return nil, err
	}
	return json.Marshal(streamInfoJSON{
		ID:              s.ID,
		LocalIdentity:   s.LocalIdentity,
		RemoteIdentity:  s.RemoteIdentity,
		LocalEndpoint:   localEP,
		RemoteEndpoint:  remoteEP,
		Outbound:        s.Outbound,
		Network:         s.Network,
		HighPressure:    s.HighPressure,
		BytesThroughput: s.BytesThroughput,
	})
}

func (s *StreamInfo) UnmarshalJSON(data []byte) error {
	var v streamInfoJSON
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	localEP, err := unmarshalEndpointJSON(v.LocalEndpoint)
	if err != nil {
		return err
	}
	remoteEP, err := unmarshalEndpointJSON(v.RemoteEndpoint)
	if err != nil {
		return err
	}
	s.ID = v.ID
	s.LocalIdentity = v.LocalIdentity
	s.RemoteIdentity = v.RemoteIdentity
	s.LocalEndpoint = localEP
	s.RemoteEndpoint = remoteEP
	s.Outbound = v.Outbound
	s.Network = v.Network
	s.HighPressure = v.HighPressure
	s.BytesThroughput = v.BytesThroughput
	return nil
}

func init() {
	astral.Add(&StreamInfo{})
}
