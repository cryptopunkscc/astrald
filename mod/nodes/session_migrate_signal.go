package nodes

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
)

// LinkSelector identifies a target link for migration signaling (Phase 0).
// Note: StreamId is local to the initiator in Phase 0; used for auditing only.
type LinkSelector struct {
	Identity *astral.Identity
	StreamId astral.Int64
}

// SessionMigrateSignal represents control-plane messages exchanged during migration signaling.
type SessionMigrateSignal struct {
	Signal astral.String8
	Nonce  astral.Nonce
}

const (
	MigrateSignalTypeBegin = "migrate_begin"
	MigrateSignalTypeReady = "migrate_ready"
)

func (m SessionMigrateSignal) ObjectType() string { return "mod.nodes.migrate_signal" }

func (m SessionMigrateSignal) WriteTo(w io.Writer) (int64, error) { return astral.Struct(m).WriteTo(w) }

func (m *SessionMigrateSignal) ReadFrom(r io.Reader) (int64, error) {
	return astral.Struct(m).ReadFrom(r)
}

func (m SessionMigrateSignal) MarshalJSON() ([]byte, error) {
	type alias SessionMigrateSignal
	return json.Marshal(alias(m))
}

func (m *SessionMigrateSignal) UnmarshalJSON(b []byte) error {
	type alias SessionMigrateSignal
	var a alias
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}
	*m = SessionMigrateSignal(a)
	return nil
}

func init() { _ = astral.Add(&SessionMigrateSignal{}) }

// ExpectMigrateSignal returns a handler that validates signal type and session nonce.
func ExpectMigrateSignal(sessionID astral.Nonce, sigType string) func(*SessionMigrateSignal) error {
	return func(sig *SessionMigrateSignal) error {
		if string(sig.Signal) != sigType {
			return fmt.Errorf("expected %s, got %s", sigType, sig.Signal)
		}
		if sig.Nonce != sessionID {
			return fmt.Errorf("session nonce mismatch")
		}
		return channel.ErrBreak
	}
}
