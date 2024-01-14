package proto

import (
	"github.com/cryptopunkscc/astrald/cslq/rpc"
	"io"
)

type Session struct {
	*rpc.Session[string]
}

func New(c io.ReadWriter) Session {
	return Session{rpc.NewSession[string](c, es)}
}

func (s *Session) Query(params *QueryParams) (*QueryResponse, error) {
	if err := s.Encodef("v", params); err != nil {
		return nil, err
	}

	if err := s.DecodeErr(); err != nil {
		return nil, err
	}

	var response QueryResponse
	if err := s.Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}
