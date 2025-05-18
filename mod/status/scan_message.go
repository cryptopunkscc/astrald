package status

import (
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

type ScanMessage struct{}

// astral

func (ScanMessage) ObjectType() string { return "mod.status.scan_message" }

func (s ScanMessage) WriteTo(w io.Writer) (n int64, err error) {
	return 0, nil
}

func (s *ScanMessage) ReadFrom(r io.Reader) (n int64, err error) {
	return 0, nil
}

func init() {
	_ = astral.DefaultBlueprints.Add(&ScanMessage{})
}
