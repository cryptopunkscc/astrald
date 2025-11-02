package objects

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/streams"
)

type SourcedObject struct {
	astral.ObjectReader
	Source *astral.Identity
	Object astral.Object
}

// astral

func (o *SourcedObject) ObjectType() string {
	return "mod.objects.sourced_object"
}

func (o *SourcedObject) WriteTo(w io.Writer) (n int64, err error) {
	return streams.WriteAllTo(w,
		o.Source,
		astral.ObjectType(o.Object.ObjectType()),
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

	o.Object, _, err = o.ObjectReader.ReadObject()
	return
}

// ...

func init() {
	_ = astral.DefaultBlueprints.Add(&SourcedObject{})
}
