package logfmt

import (
	"github.com/cryptopunkscc/astrald/auth/id"
)

func ID(i id.Identity) string {
	s := i.PublicKeyHex()
	return s[len(s)-8:len(s)-4] + ":" + s[len(s)-4:]
}

func Dir(out bool) string {
	if out {
		return "out"
	}
	return "in"
}

func Bool(b bool, true string, false string) string {
	if b {
		return true
	}
	return false
}
