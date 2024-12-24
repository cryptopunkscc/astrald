package arl

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"regexp"
	"strings"
)

var callerExp = regexp.MustCompile(`^([a-zA-Z0-9_.-]+)@`)
var queryExp = regexp.MustCompile(`^[a-zA-Z0-9_.-]+:(.*)$`)

// ARL - Astral Resource Locator
type ARL struct {
	Caller *astral.Identity
	Target *astral.Identity
	Query  string
}

func New(caller *astral.Identity, target *astral.Identity, query string) *ARL {
	return &ARL{Caller: caller, Target: target, Query: query}
}

func Split(s string) (caller, target, query string) {
	matches := callerExp.FindStringSubmatch(s)
	if len(matches) > 0 {
		s, _ = strings.CutPrefix(s, matches[0])
		caller = matches[1]
	}

	matches = queryExp.FindStringSubmatch(s)
	if len(matches) > 0 {
		s, _ = strings.CutSuffix(s, ":"+matches[1])
		query = matches[1]
	}

	target = s
	return
}

func Parse(s string, resolver dir.Resolver) (arl *ARL, err error) {
	if after, found := strings.CutPrefix(s, "astral://"); found {
		s = after
	}

	var c, t string
	arl = &ARL{}
	c, t, arl.Query = Split(s)

	if len(c) != 0 {
		if resolver != nil {
			arl.Caller, err = resolver.ResolveIdentity(c)
			if err != nil {
				return
			}
		} else {
			arl.Caller, err = astral.IdentityFromString(c)
			if err != nil {
				return
			}
		}
	}

	if len(t) != 0 {
		if resolver != nil {
			arl.Target, err = resolver.ResolveIdentity(t)
			if err != nil {
				return
			}
		} else {
			arl.Target, err = astral.IdentityFromString(t)
			if err != nil {
				return
			}
		}
	}

	return
}

func (arl *ARL) String() (s string) {
	if !arl.Caller.IsZero() {
		s = arl.Caller.String() + "@"
	}
	if !arl.Target.IsZero() {
		s = s + arl.Target.String() + ":"
	}
	if len(arl.Query) > 0 {
		s = s + arl.Query
	}
	return
}
