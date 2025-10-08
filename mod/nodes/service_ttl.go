package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

type ServiceTTL struct {
	ProviderID *astral.Identity // can be the local node or an app
	Name       astral.String8
	Priority   astral.Uint16
	TTL        astral.Duration // time to live
}

func (ServiceTTL) ObjectType() string { return "nodes.service_ttl" }

func (s ServiceTTL) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(s).WriteTo(w)
}

func (s *ServiceTTL) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(s).ReadFrom(r)
}

func init() {
	_ = astral.DefaultBlueprints.Add(&ServiceTTL{})
}
