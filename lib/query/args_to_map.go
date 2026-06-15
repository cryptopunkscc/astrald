package query

import "strings"

// ArgsToMap converts a flag-style argument slice to a map. Arguments prefixed
// with "-" are treated as named keys consuming the next element as their value;
// unprefixed arguments are stored under DefaultArgKey.
func ArgsToMap(args []string) (params map[string]string) {
	params = make(map[string]string)

	for len(args) > 0 {
		key := args[0]

		if !strings.HasPrefix(key, "-") {
			params[DefaultArgKey] = key
			args = args[1:]
			continue
		}

		key = key[1:]

		if len(args) < 2 {
			params[key] = ""
			return
		}

		params[key] = args[1]
		args = args[2:]
	}
	return
}
