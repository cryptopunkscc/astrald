package term

import "unicode"

func ToSnakeCase(str string) string {
	var result []rune
	var lastUpper bool
	for i, r := range str {
		if unicode.IsUpper(r) {
			if i > 0 {
				if (!lastUpper && !unicode.IsDigit(r)) || (i+1 < len(str) && !unicode.IsUpper(rune(str[i+1])) && !unicode.IsDigit(rune(str[i+1]))) {
					result = append(result, '_')
				}
			}
			result = append(result, unicode.ToLower(r))
			lastUpper = true
		} else {
			result = append(result, r)
			lastUpper = unicode.IsUpper(r) || unicode.IsDigit(r)
		}
	}
	return string(result)
}
