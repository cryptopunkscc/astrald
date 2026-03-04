package nodes

import (
	"bytes"
	"encoding"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
)

type SessionInfo struct {
	ID             astral.Nonce
	StreamID       astral.Nonce
	RemoteIdentity *astral.Identity
	Outbound       astral.Bool
	Query          astral.String16
	Bytes          astral.Uint64
	Age            astral.Duration
	CanMigrate     astral.Bool
}

var _ astral.Object = &SessionInfo{}
var _ encoding.TextMarshaler = &SessionInfo{}
var _ json.Marshaler = &SessionInfo{}

func (s SessionInfo) ObjectType() string { return "mod.nodes.session_info" }

func (s SessionInfo) WriteTo(w io.Writer) (int64, error) {
	return astral.Objectify(&s).WriteTo(w)
}

func (s *SessionInfo) ReadFrom(r io.Reader) (int64, error) {
	return astral.Objectify(s).ReadFrom(r)
}

func (s SessionInfo) MarshalText() ([]byte, error) {
	var b bytes.Buffer
	d := "<"
	if s.Outbound {
		d = ">"
	}
	migrate := ""
	if s.CanMigrate {
		migrate = " [migratable]"
	}
	age := time.Duration(s.Age).Round(time.Second)
	_, err := fmt.Fprintf(&b, "%v stream=%v %v %v %v bytes=%v age=%v%v",
		s.ID, s.StreamID, d, s.RemoteIdentity, s.Query, s.Bytes, age, migrate)
	return b.Bytes(), err
}

func (s SessionInfo) MarshalJSON() ([]byte, error) {
	type Alias SessionInfo
	return json.Marshal(Alias(s))
}

func init() {
	astral.Add(&SessionInfo{})
}
