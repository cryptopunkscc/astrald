package nodes

import (
	"fmt"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
)

// MigrateSignal is exchanged over the signalling channel during session migration.
type MigrateSignal struct {
	Signal astral.String8
}

const (
	MigrateSignalReady    astral.String8 = "ready"
	MigrateSignalSwitched astral.String8 = "switched"
	MigrateSignalResume   astral.String8 = "resume"
)

func (MigrateSignal) ObjectType() string { return "mod.nodes.migrate_signal" }

func (m MigrateSignal) WriteTo(w io.Writer) (int64, error) {
	return astral.Struct(m).WriteTo(w)
}

func (m *MigrateSignal) ReadFrom(r io.Reader) (int64, error) {
	return astral.Struct(m).ReadFrom(r)
}

func init() { _ = astral.Add(&MigrateSignal{}) }

// ExpectMigrateSignal returns a channel.Switch handler that accepts the given signal type.
func ExpectMigrateSignal(signalType astral.String8) func(*MigrateSignal) error {
	return func(sig *MigrateSignal) error {
		if sig.Signal != signalType {
			return fmt.Errorf("expected %s signal, got %s", signalType, sig.Signal)
		}
		return channel.ErrBreak
	}
}
