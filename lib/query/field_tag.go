package query

import "strings"

type FieldTag struct {
	Key      string
	Skip     bool
	Required bool
	Other    map[string]string
}

func ParseTag(tag string) *FieldTag {
	var fieldTag = FieldTag{Other: make(map[string]string)}

	s := strings.Split(tag, ";")
	for _, v := range s {
		p := strings.SplitN(v, ":", 2)

		switch p[0] {
		case "skip":
			fieldTag.Skip = true
		case "required":
			fieldTag.Required = true
		case "key":
			if len(p) < 2 {
				continue
			}
			fieldTag.Key = p[1]
		default:
			if len(p) == 2 {
				fieldTag.Other[p[0]] = p[1]
			} else {
				fieldTag.Other[p[0]] = ""
			}
		}
	}

	return &fieldTag
}
