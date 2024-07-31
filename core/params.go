package core

import (
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/object"
	"strconv"
	"strings"
	"time"
)

var ErrKeyNotFound = errors.New("key not found")

type Params map[string]string

func (params Params) GetInt(key string) (int, error) {
	v, found := params[key]
	if !found {
		return 0, ErrKeyNotFound
	}

	i, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse error: %w", err)
	}

	return int(i), nil
}

func (params Params) GetUint64(key string) (uint64, error) {
	v, found := params[key]
	if !found {
		return 0, ErrKeyNotFound
	}

	i, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse error: %w", err)
	}

	return i, nil
}

func (params Params) SetInt(key string, value int) {
	params[key] = strconv.FormatInt(int64(value), 10)
}

func (params Params) GetObjectID(key string) (object.ID, error) {
	v, found := params[key]
	if !found {
		return object.ID{}, ErrKeyNotFound
	}

	id, err := object.ParseID(v)
	if err != nil {
		return object.ID{}, err
	}

	return id, nil
}

func (params Params) GetUnixNano(key string) (time.Time, error) {
	v, found := params[key]
	if !found {
		return time.Time{}, ErrKeyNotFound
	}

	nsec, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse error: %w", err)
	}

	return time.Unix(0, nsec), nil
}

func (params Params) SetUnixNano(key string, value time.Time) {
	params[key] = strconv.FormatInt(value.UnixNano(), 10)
}

func SplitQueryParams(query string) (path, params string) {
	if i := strings.IndexByte(query, '?'); i != -1 {
		return query[:i], query[i+1:]
	}
	return query, ""
}

func ParseParams(params string) Params {
	var p = map[string]string{}

	var list = strings.Split(params, "&")
	for _, item := range list {
		var key, value string
		s := strings.SplitN(item, "=", 2)
		if len(s) == 2 {
			key, value = s[0], s[1]
		} else {
			value = s[0]
		}
		p[key] = value
	}

	return p
}

func ParseQuery(query string) (path string, params Params) {
	var s string
	path, s = SplitQueryParams(query)
	params = ParseParams(s)
	return
}

func Query(path string, params Params) string {
	var f = path
	var l []string
	for k, v := range params {
		l = append(l, k+"="+v)
	}
	if len(l) > 0 {
		f = f + "?" + strings.Join(l, "&")
	}

	return f
}
