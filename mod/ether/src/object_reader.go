package ether

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"io"
)

var _ astral.ObjectReader = &objectReader{}

type objectReader struct {
	io.Reader
	objects objects.Module
}

func (o objectReader) ReadObject(r io.Reader) (astral.Object, error) {
	return o.objects.ReadObject(r)
}
