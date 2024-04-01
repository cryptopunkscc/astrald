package storage

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	proto "github.com/cryptopunkscc/astrald/mod/storage/srv"
	"io"
)

type IdentityFilterService struct {
	id.Filter
	port string
}

func NewIdentityFilterService(filter id.Filter) (s *IdentityFilterService, err error) {
	var portId id.Identity
	if portId, err = id.GenerateIdentity(); err != nil {
		return
	}
	return &IdentityFilterService{Filter: filter, port: portId.String()}, nil
}

func (s *IdentityFilterService) String() string {
	return s.port
}

func (s *IdentityFilterService) Handle(conn io.ReadWriter, _ string) (err error) {
	enc := proto.NewBinaryEncoder(conn)
	for {
		var identity id.Identity
		bytes := make([]byte, 33)
		if err = enc.Decode(&bytes); err != nil {
			return
		}
		if identity, err = id.ParsePublicKey(bytes); err != nil {
			return
		}
		b := s.Filter(identity)
		if err = enc.Encode(b); err != nil {
			return
		}
	}
}
