package core

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/sig"
	"reflect"
	"strings"
)

var _ node.AuthEngine = &CoreAuthorizer{}

type CoreAuthorizer struct {
	authorizers sig.Set[node.Authorizer]
	log         *log.Logger
}

func NewCoreAuthorizer(log *log.Logger) (*CoreAuthorizer, error) {
	return &CoreAuthorizer{log: log}, nil
}

func (auth *CoreAuthorizer) Authorize(identity id.Identity, action string, args ...any) bool {
	for _, a := range auth.authorizers.Clone() {
		if a.Authorize(identity, action, args...) {
			name := reflect.TypeOf(a).String()
			if s, ok := a.(fmt.Stringer); ok {
				name = s.String()
			}

			var fmt = "allowed %v to %v" + strings.Repeat(" %v", len(args)) + " by %v"
			var vals = []any{identity, action}
			vals = append(vals, args...)
			vals = append(vals, name)

			auth.log.Infov(2, fmt, vals...)
			return true
		}
	}

	auth.log.Logv(2, "denied %v to %v", identity, action)

	return false
}

func (auth *CoreAuthorizer) Add(authorizer node.Authorizer) error {
	return auth.authorizers.Add(authorizer)
}

func (auth *CoreAuthorizer) Remove(authorizer node.Authorizer) error {
	return auth.authorizers.Remove(authorizer)
}
