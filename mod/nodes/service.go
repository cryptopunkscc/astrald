package nodes

import (
	"encoding/json"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

type Service struct {
	ProviderID *astral.Identity
	Name       astral.String8
	Priority   astral.Uint16
}

// astral

func (Service) ObjectType() string { return "nodes.service" }

func (s Service) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(s).WriteTo(w)
}

func (s *Service) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(s).ReadFrom(r)
}

// json

func (s Service) MarshalJSON() ([]byte, error) {
	type alias Service
	return json.Marshal(alias(s))
}

func (s *Service) UnmarshalJSON(bytes []byte) error {
	type alias Service
	var a alias

	err := json.Unmarshal(bytes, &a)
	if err != nil {
		return err
	}

	*s = Service(a)
	return nil
}

// text

func (s Service) MarshalText() (text []byte, err error) {
	text = []byte(fmt.Sprintf("%s/%s prio %d", s.ProviderID, s.Name, s.Priority))
	return
}

func init() {
	_ = astral.DefaultBlueprints.Add(&Service{})
}
