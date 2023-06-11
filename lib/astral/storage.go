package astral

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/storage/rpc"
	"io"
)

type Storage struct {
	client   *ApphostClient
	Identity id.Identity
	Service  string
}

func (s *Storage) Read(dataID data.ID, start int, n int) (r io.Reader, err error) {
	conn, err := s.client.Query(s.Identity, s.Service)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	return rpc.New(conn).Read(dataID, start, n)
}
