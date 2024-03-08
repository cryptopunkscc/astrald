package storage

import (
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"io"
)

type CreatorService struct {
	storage.Creator
	port string
}

func NewCreatorService(creator storage.Creator) *CreatorService {
	return &CreatorService{Creator: creator}
}

func (s *CreatorService) Port(port string) *CreatorService {
	s.port = port
	return s
}

func (s *CreatorService) String() string {
	if s.port == "" {
		panic("port not set")
	}
	return s.port
}

func (s *CreatorService) Handle(conn io.ReadWriter, args *cslq.Decoder) (err error) {
	var opts storage.CreateOpts
	if err = args.Decodef("v", &opts); err != nil {
		return
	}
	creator, err := s.Creator.Create(&opts)
	if err != nil {
		return
	}
	if err = newWriterService(creator, conn).Loop(); err != nil {
		return
	}
	return
}
