package storage

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"io"
)

type OpenerService struct {
	storage.Opener
	port     string
	remoteID id.Identity
}

func NewOpenerService(opener storage.Opener, remoteID id.Identity) *OpenerService {
	return &OpenerService{Opener: opener, remoteID: remoteID}
}

func (s *OpenerService) String() string {
	if s.port == "" {
		panic("port not set")
	}
	return s.port
}

func (s *OpenerService) Port(port string) *OpenerService {
	s.port = port
	return s
}

func (s *OpenerService) Handle(conn io.ReadWriter, args *cslq.Decoder) (err error) {
	// decode args
	var dataID data.ID
	var opts storage.OpenOpts
	var idFilterPort string
	if err = args.Decodef("v v [c]c", &dataID, &opts, &idFilterPort); err != nil {
		return
	}

	// inject id filter client if port was specified
	var closer io.Closer
	defer func() {
		if closer != nil {
			_ = closer.Close()
		}
	}()
	if idFilterPort != "" {
		if opts.IdentityFilter, closer, err = NewTargetClient(s.remoteID, idFilterPort).idFilter(); err != nil {
			return
		}
	}

	// handle requests in loop
	reader, err := s.Open(dataID, &opts)
	if err != nil {
		return
	}
	if err = newReaderService(reader, conn).Loop(); err != nil {
		return
	}
	return
}
