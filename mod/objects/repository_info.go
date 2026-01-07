package objects

import (
	"encoding/json"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

type RepositoryInfo struct {
	Name  astral.String8
	Label astral.String8
	Free  astral.Int64
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

// ...

func init() {
	_ = astral.DefaultBlueprints.Add(&RepositoryInfo{})
}
