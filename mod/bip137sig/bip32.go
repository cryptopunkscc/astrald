package bip137sig

import (
	"fmt"
	"strconv"
	"strings"
)

// reference: https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki

func ParseDerivationPath(path string) ([]uint32, error) {
	const hardenedOffset uint32 = 0x80000000

	if path == "m" || path == "" {
		return nil, nil
	}

	if strings.HasPrefix(path, "m/") {
		path = path[2:]
	}

	parts := strings.Split(path, "/")
	out := make([]uint32, 0, len(parts))

	for _, p := range parts {
		hardened := false

		if strings.HasSuffix(p, "'") || strings.HasSuffix(p, "h") {
			hardened = true
			p = p[:len(p)-1]
		}

		i, err := strconv.ParseUint(p, 10, 31)
		if err != nil {
			return nil, fmt.Errorf("invalid path element %q", p)
		}

		if hardened {
			i |= uint64(hardenedOffset)
		}

		out = append(out, uint32(i))
	}

	return out, nil
}
