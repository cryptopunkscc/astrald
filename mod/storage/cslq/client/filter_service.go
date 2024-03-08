package storage

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"io"
)

type identityFilterService struct {
	id.Filter
	port string
}

func newIdentityFilterService(filter id.Filter) (s *identityFilterService, err error) {
	var portId id.Identity
	if portId, err = id.GenerateIdentity(); err != nil {
		return
	}
	return &identityFilterService{Filter: filter, port: portId.String()}, nil
}

func (s *identityFilterService) String() string {
	return s.port
}

func (s *identityFilterService) Handle(conn io.ReadWriter, _ *cslq.Decoder) (err error) {
	var identity id.Identity
	if err = cslq.Decode(conn, "v", &identity); err != nil {
		return
	}
	b := s.Filter(identity)
	if err = cslq.Encode(conn, "c", b); err != nil {
		return
	}
	return
}
