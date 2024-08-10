package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/streams"
	"io"
)

var _ astral.Object = &SourcedObject{}

type SourcedObject struct {
	ObjectReader
	Source *astral.Identity
	Object astral.Object
}

func (o *SourcedObject) ObjectType() string {
	return "astrald.mod.objects.sourced_object"
}

func (o *SourcedObject) WriteTo(w io.Writer) (n int64, err error) {
	return streams.WriteAllTo(w,
		o.Source,
		astral.ObjectHeader(o.Object.ObjectType()),
		o.Object)
}

func (o *SourcedObject) ReadFrom(r io.Reader) (n int64, err error) {
	c := streams.NewReadCounter(r)
	defer func() {
		n = c.Total()
	}()

	o.Source = &astral.Identity{}
	_, err = o.Source.ReadFrom(c)
	if err != nil {
		return
	}

	o.Object, err = o.ObjectReader.ReadObject(c)
	return
}
