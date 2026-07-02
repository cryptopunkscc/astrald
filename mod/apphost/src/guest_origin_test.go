package apphost

import (
	"testing"

	"github.com/cryptopunkscc/astrald/astral"
)

// TestWebOriginDenied covers the transport origin gate: an anonymous guest with a
// browser Origin may route only from the first-party origin; a token bypasses the
// check from any origin; a non-browser guest (empty Origin) is never denied here.
func TestWebOriginDenied(t *testing.T) {
	authed := astral.GenerateIdentity()

	cases := []struct {
		name      string
		guestID   *astral.Identity
		webOrigin string
		want      bool
	}{
		{"anon, no origin (IPC/non-browser)", nil, "", false},
		{"anon, first-party origin", nil, TrustedWebOrigin, false},
		{"anon, untrusted origin", nil, "https://evil.example", true},
		{"authenticated, untrusted origin", authed, "https://evil.example", false},
		{"authenticated, first-party origin", authed, TrustedWebOrigin, false},
		{"authenticated, no origin", authed, "", false},
	}

	mod := &Module{config: Config{TrustedWebOrigins: []string{TrustedWebOrigin}}}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			guest := &Guest{mod: mod, guestID: c.guestID, webOrigin: c.webOrigin}
			if got := guest.webOriginDenied(); got != c.want {
				t.Fatalf("webOriginDenied() = %v; want %v", got, c.want)
			}
		})
	}
}
