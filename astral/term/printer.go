package term

import (
	"github.com/cryptopunkscc/astrald/astral"
)

type Printer interface {
	Print(...astral.Object) error
}

type PrinterTo interface {
	PrintTo(Printer) error
}
