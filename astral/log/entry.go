package log

import (
	"io"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
)

var _ astral.Object = &Entry{}

type Entry struct {
	Origin  *astral.Identity
	Level   uint8
	Time    astral.Time
	Objects []astral.Object
}

func NewEntry(origin *astral.Identity, level uint8, obj ...astral.Object) *Entry {
	return &Entry{
		Origin:  origin,
		Level:   level,
		Time:    astral.Time(time.Now()),
		Objects: obj,
	}
}

func (Entry) ObjectType() string {
	return "astrald.log.entry"
}

func (e Entry) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Objectify(&e).WriteTo(w)
}

func (e *Entry) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Objectify(e).ReadFrom(r)
}

func (e Entry) MarshalJSON() ([]byte, error) {
	return astral.Objectify(&e).MarshalJSON()
}

func (e *Entry) UnmarshalJSON(bytes []byte) error {
	return astral.Objectify(e).UnmarshalJSON(bytes)
}

func init() {
	astral.DefaultBlueprints.Add(&Entry{})
}
