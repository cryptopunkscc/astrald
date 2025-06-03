package query

import (
	"github.com/cryptopunkscc/astrald/astral"
	"net/url"
	"strings"
	"time"
)

const DefaultArgKey = "arg"
const queryTag = "query"
const maxQueryTimeout = 60 * time.Second

type Validator interface {
	Validate() error
}

type Args map[string]any

func New(caller *astral.Identity, target *astral.Identity, path string, args any) (query *astral.Query) {
	query = &astral.Query{
		Nonce:  astral.NewNonce(),
		Caller: caller,
		Target: target,
		Query:  path,
	}

	if args == nil {
		return
	}

	str, err := Marshal(args)
	if err != nil {
		return
	}

	if len(str) > 0 {
		query.Query += "?" + str
	}

	return
}

func Parse(q string) (path string, params map[string]string) {
	var s string
	path, s = splitPathParams(q)
	params = map[string]string{}

	vals, err := url.ParseQuery(s)
	if err != nil {
		return
	}

	for k, v := range vals {
		if len(v) > 0 {
			params[k] = v[0]
		} else {
			params[DefaultArgKey] = k
		}
	}

	return
}

func splitTag(tag string) (m map[string]string) {
	m = make(map[string]string)

	s := strings.Split(tag, ";")
	for _, v := range s {
		p := strings.SplitN(v, ":", 2)
		if len(p) < 2 {
			m[p[0]] = ""
		} else {
			m[p[0]] = p[1]
		}
	}

	return m
}

func splitPathParams(query string) (path, params string) {
	if i := strings.IndexByte(query, '?'); i != -1 {
		return query[:i], query[i+1:]
	}
	return query, ""
}
