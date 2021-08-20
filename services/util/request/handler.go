package request

import (
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/sio"
)

type Handler func(
	caller api.Identity,
	query string,
	stream sio.ReadWriteCloser,
) error
