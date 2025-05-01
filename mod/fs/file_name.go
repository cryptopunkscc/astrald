package fs

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

type FileName astral.String16

// astral

func (fn FileName) ObjectType() string {
	return "mod.fs.file_name"
}

func (fn FileName) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(fn).WriteTo(w)
}

func (fn *FileName) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(fn).ReadFrom(r)
}

// json

func (fn FileName) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(fn))
}

func (fn *FileName) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, (*string)(fn))
}

// text

func (fn FileName) MarshalText() (text []byte, err error) {
	return []byte(fn), nil
}

func (fn *FileName) UnmarshalText(text []byte) error {
	*fn = FileName(text)
	return nil
}

// other

func (fn FileName) String() string {
	return string(fn)
}

func init() {
	var fn FileName
	astral.DefaultBlueprints.Add(&fn)
}
