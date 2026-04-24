package fmt

import "github.com/cryptopunkscc/astrald/astral"

func Format(format string, args ...any) (list []any) {
	var token string

	for len(format) > 0 {
		token, format = cutToken(format)

		if len(token) <= 1 {
			list = append(list, astral.NewString32(token))
			continue
		}

		switch token[0] {
		case '%':
			switch token[1] {
			case 'v':
				if len(args) == 0 {
					list = append(list, astral.NewError("#[err_arg_missing]"))
					continue
				}

				var a any
				a, args = args[0], args[1:]

				list = append(list, a)

			case 's', 'd':
				if len(args) == 0 {
					list = append(list, astral.NewError("#[err_arg_missing]"))
					continue
				}

				var a any
				a, args = args[0], args[1:]

				list = append(list, astral.Stringify(a))

			default:
				list = append(list, astral.NewString32(token))
			}

		case '\\':
			switch token[1] {
			case 't':
				list = append(list, astral.NewString32("\t"))

			case 'n':
				list = append(list, astral.NewString32("\n"))

			case 'r':
				list = append(list, astral.NewString32("\r"))

			default:
				list = append(list, astral.NewString32(token))

			}

		default:
			list = append(list, astral.NewString32(token))
		}
	}
	return
}
