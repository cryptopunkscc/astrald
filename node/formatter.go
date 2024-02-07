package node

import "github.com/cryptopunkscc/astrald/auth/id"

type Formatter func(Node, string) string

var formatters = []Formatter{}

func AddFormatter(fn Formatter) {
	formatters = append(formatters, fn)
}

func init() {
	AddFormatter(func(node Node, s string) string {
		if len(s) != 66 {
			return ""
		}
		identity, err := id.ParsePublicKeyHex(s)
		if err != nil {
			return ""
		}

		return node.Resolver().DisplayName(identity)
	})
}
