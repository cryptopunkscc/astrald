package query

import "strings"

func ParseArgs(args []string) (params map[string]string) {
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
