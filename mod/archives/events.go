package archives

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/object"
	"io"
)

var _ astral.Object = &EventArchiveIndexed{}

type EventArchiveIndexed struct {
	ObjectID *object.ID
	Archive  *Archive
}

func (EventArchiveIndexed) ObjectType() string {
	return "astrald.mod.archives.events.archive_indexed"
}

func (e EventArchiveIndexed) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(e).WriteTo(w)
}

func (e *EventArchiveIndexed) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(e).ReadFrom(r)
}
