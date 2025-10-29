package nodes

import (
	"encoding/json"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
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
	MigrateSignalTypeBegin     = "migrate_begin"
	MigrateSignalTypeReady     = "migrate_ready"
	MigrateSignalTypeCompleted = "migrate_completed"
	MigrateSignalTypeAbort     = "migrate_abort"
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

func init() { _ = astral.DefaultBlueprints.Add(&SessionMigrateSignal{}) }
