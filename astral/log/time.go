package log

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/term"
	"time"
)

type Time astral.Time

const shortTimestamp = "15:04:05.000"

func (t Time) PrintTo(p term.Printer) error {
	var s = astral.String(time.Time(t).Local().Format(shortTimestamp))

	return p.Print(&s)
}
