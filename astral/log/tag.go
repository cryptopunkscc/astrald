package log

import (
	"encoding/json"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

type Tag string

func (Tag) ObjectType() string { return "astrald.log.tag" }

func (l Tag) WriteTo(w io.Writer) (n int64, err error) {
	return astral.String8(l).WriteTo(w)
}

func (l *Tag) ReadFrom(r io.Reader) (n int64, err error) {
	return (*astral.String8)(l).ReadFrom(r)
}

func (l Tag) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(l))
}

func (l Tag) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, (*string)(&l))
}

func (l Tag) Render() string { return "[" + string(l) + "] " }

func (l Tag) String() string {
	return string(l)
}

func init() {
	var v Tag
	astral.Add(&v)
}
