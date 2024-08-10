package content

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/cslq"
	"io"
)

var _ astral.Object = &ObjectDescriptor{}

type ObjectDescriptor struct {
	Type string
}

func (d *ObjectDescriptor) ObjectType() string {
	return "astrald.mod.content.object_type"
}

func (d *ObjectDescriptor) WriteTo(w io.Writer) (n int64, err error) {
	err = cslq.Encode(w, "[c]c", d.Type)
	if err == nil {
		n = int64(len(d.Type) + 1)
	}
	return
}

func (d *ObjectDescriptor) ReadFrom(r io.Reader) (n int64, err error) {
	err = cslq.Decode(r, "[c]c", &d.Type)
	if err == nil {
		n = int64(len(d.Type) + 1)
	}
	return
}
