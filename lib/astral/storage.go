package astral

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/storage/proto"
	"io"
)

type Storage struct {
	client   *ClientInfo
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

	stream := proto.New(conn)

	err = stream.Encode(proto.MsgRead{
		DataID: dataID,
		Start:  int64(start),
		Len:    int64(n),
	})
	if err != nil {
		return nil, err
	}

	if err = stream.ReadError(); err != nil {
		return nil, err
	}

	return conn, nil
}
