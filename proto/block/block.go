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
		io.Closer
		Finalizer
	}

	Finalizer interface {
		Finalize() (data.ID, error)
	}
)
