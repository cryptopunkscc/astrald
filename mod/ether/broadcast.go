package ether

import (
	"bytes"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/streams"
	"io"
)

type ObjectReader interface {
	ReadObject() (astral.Object, error)
}

type Broadcast struct {
	Object    astral.Object
	Timestamp astral.Time
	Source    *astral.Identity
}

var _ astral.Object = &Broadcast{}

func (Broadcast) ObjectType() string { return "astrald.mod.ether.broadcast" }

func (p Broadcast) WriteTo(w io.Writer) (n int64, err error) {
	var payload = &bytes.Buffer{}
	objects.WriteObject(payload, p.Object)

	return streams.WriteAllTo(w,
		astral.Bytes16(payload.Bytes()),
		p.Timestamp,
		p.Source)
}

func (p *Broadcast) ReadFrom(r io.Reader) (n int64, err error) {
	var payload astral.Bytes16
	p.Source = &astral.Identity{}

	n, err = streams.ReadAllFrom(r,
		&payload,
		&p.Timestamp,
		p.Source)
	if err != nil {
		return
	}

	pr := bytes.NewReader(payload)

	if or, ok := r.(astral.ObjectReader); ok {
		p.Object, err = or.ReadObject(pr)
	} else {
		var h astral.ObjectHeader
		_, err = h.ReadFrom(pr)
		if err != nil {
			return
		}

		var obj = &objects.ForeignObject{Type: string(h)}
		_, err = obj.ReadFrom(pr)
		if err != nil {
			return
		}

		p.Object = obj
	}

	return
}
