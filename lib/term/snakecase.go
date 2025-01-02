package term

import "unicode"

func ToSnakeCase(str string) string {
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
