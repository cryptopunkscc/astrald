package query

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"net/url"
	"strings"
	"time"
	"unicode"
)

const queryTag = "query"
const defaultQueryTimeout = 60 * time.Second

func New(caller *astral.Identity, target *astral.Identity, path string, args any) *astral.Query {
	q, err := Marshal(args)
	if err != nil {
		return nil
	}
	return &astral.Query{
		Nonce:  astral.NewNonce(),
		Caller: caller,
		Target: target,
		Query:  path + "?" + q,
	}
}

func Run(n astral.Node, target *astral.Identity, path string, args any) (astral.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultQueryTimeout)
	defer cancel()

	return RunCtx(ctx, n, target, path, args)
}

func RunCtx(ctx context.Context, n astral.Node, target *astral.Identity, path string, args any) (astral.Conn, error) {
	q := New(n.Identity(), target, path, args)

	return Route(ctx, n, q)
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
			params[k] = ""
		}
	}

	return
}

func ParseTo(q string, args any) (path string, err error) {
	path, params := Parse(q)
	err = populate(params, args)
	return
}

func toSnakeCase(str string) string {
	var result []rune
	var lastUpper bool
	for i, r := range str {
		if unicode.IsUpper(r) {
			if i > 0 && !lastUpper {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
			lastUpper = true
		} else {
			result = append(result, r)
			lastUpper = false
		}
	}
	return string(result)
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
