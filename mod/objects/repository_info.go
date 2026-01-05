package objects

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

type RepositoryInfo struct {
	ID    astral.String8
	Label astral.String8
	Free  astral.Uint64
}

var _ astral.Object = &RepositoryInfo{}

// astral

func (RepositoryInfo) ObjectType() string {
	return "mod.objects.repository_info"
}

func (info RepositoryInfo) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(info).WriteTo(w)
}

func (info *RepositoryInfo) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(info).ReadFrom(r)
}

// json

func (info RepositoryInfo) MarshalJSON() ([]byte, error) {
	type alias RepositoryInfo
	return json.Marshal(alias(info))
}

func (info *RepositoryInfo) UnmarshalJSON(bytes []byte) error {
	type alias RepositoryInfo
	return json.Unmarshal(bytes, (*alias)(info))
}

// text

func (info RepositoryInfo) MarshalText() (text []byte, err error) {
	return []byte(fmt.Sprintf(
		"%s: %s (%s free)",
		info.ID,
		info.Label,
		astral.Size(info.Free).HumanReadable(),
	)), nil
}

// ...

func init() {
	_ = astral.DefaultBlueprints.Add(&RepositoryInfo{})
}
