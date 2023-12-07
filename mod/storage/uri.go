package storage

import (
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"strings"
)

type URI struct {
	User   id.Identity
	Target id.Identity
	Query  string
}

var userPattern = "^[a-zA-Z0-9]([a-zA-Z0-9_-]*[a-zA-Z0-9])?$"

func (mod *Module) Parse(s string) (*URI, error) {
	var uri URI
	var err error

	if idx := strings.IndexByte(s, '@'); idx != -1 {
		user := s[:idx]
		uri.User, err = mod.node.Resolver().Resolve(user)
		if err != nil {
			return nil, err
		}
		s = s[idx+1:]
	}

	var idx = strings.IndexByte(s, ':')
	if idx == -1 {
		return nil, errors.New("missing query part")
	}

	var target = s[:idx]
	uri.Target, err = mod.node.Resolver().Resolve(target)
	if err != nil {
		return nil, err
	}

	uri.Query = s[idx+1:]

	return &uri, nil
}
