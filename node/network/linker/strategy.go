package linker

import "github.com/cryptopunkscc/astrald/astral/link"

type Strategy interface {
	Wake()
	Links() <-chan *link.Link
}
