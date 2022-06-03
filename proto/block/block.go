package block

import (
	"github.com/cryptopunkscc/astrald/data"
	"io"
)

type (
	// Block interface contains every io:block protocol method
	Block interface {
		io.Reader
		io.Writer
		io.Seeker
		Finalizer
		Ender
	}

	Finalizer interface {
		Finalize() (data.ID, error)
	}

	Ender interface {
		End() error
	}
)
