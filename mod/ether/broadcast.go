package ether

import (
	"bytes"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/streams"
	"io"
)

var _ astral.Object = &Broadcast{}

type Broadcast struct {
	Timestamp astral.Time
	Source    *astral.Identity
	Object    astral.Object
}

var _ astral.Object = &Broadcast{}

func (Broadcast) ObjectType() string { return "astrald.mod.ether.broadcast" }

func (b Broadcast) WriteTo(w io.Writer) (n int64, err error) {
	var buf = &bytes.Buffer{}
	astral.Write(buf, b.Object)

	return streams.WriteAllTo(w,
		b.Timestamp,
		b.Source,
		astral.Bytes16(buf.Bytes()),
	)
}

func (b *Broadcast) ReadFrom(r io.Reader) (n int64, err error) {
	var buf astral.Bytes16
	b.Source = &astral.Identity{}

	n, err = streams.ReadAllFrom(r, &b.Timestamp, b.Source, &buf)
	if err != nil {
		return
	}

	bp := astral.ExtractBlueprints(r)
	b.Object, _, err = bp.Read(bytes.NewReader(buf), false)
	if err != nil {
		return
	}

	return
}

func init() {
	astral.DefaultBlueprints.Add(&Broadcast{})
}
