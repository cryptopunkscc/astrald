package frames

import (
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
)

type Frame interface {
	astral.Object
	fmt.Stringer
}

func init() {
	_ = astral.Add(
		&Ping{},
		&Query{},
		&Response{},
		&Read{},
		&Data{},
		&Migrate{},
		&Reset{},
	)
}
