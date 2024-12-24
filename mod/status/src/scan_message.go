package status

import (
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

var _ astral.Object = &ScanMessage{}

type ScanMessage struct{}

func (ScanMessage) ObjectType() string { return "astrald.mod.presence.scan_message" }

func (s ScanMessage) WriteTo(w io.Writer) (n int64, err error) {
	return 0, nil
}

func (s *ScanMessage) ReadFrom(r io.Reader) (n int64, err error) {
	return 0, nil
}
