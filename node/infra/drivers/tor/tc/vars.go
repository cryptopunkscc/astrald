package tc

import "strings"

type VarMap map[string]string

func (m VarMap) GetList(key string) []string {
	if v, ok := m[key]; ok {
		return strings.Split(v, ",")
	}
	return []string{}
}

func (m VarMap) GetString(key string) string {
	if v, ok := m[key]; ok {
		if (len(v) > 1) && (v[0] == '"') && (v[len(v)-1] == '"') {
			return v[1 : len(v)-1]
		}
		return v
	}
	return ""
}

func parseVarMap(words []string) VarMap {
	m := make(map[string]string)
	for _, w := range words {
		kv := strings.SplitN(w, "=", 2)
		m[kv[0]] = kv[1]
	}
	return m
}
