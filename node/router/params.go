package router

import "strings"

func SplitParams(query string) (path, params string) {
	if i := strings.IndexByte(query, '?'); i != -1 {
		return query[:i], query[i+1:]
	}
	return query, ""
}

func ParseParams(params string) map[string]string {
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

func ParseQuery(query string) (path string, params map[string]string) {
	var s string
	path, s = SplitParams(query)
	params = ParseParams(s)
	return
}

func FormatQuery(path string, params map[string]string) string {
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
