package apphost

import (
	"strings"
	"testing"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
)

// TestRandomString checks the token generator's output shape and that samples do
// not collide - a weak proxy for "not predictable", enough to catch a regression
// back to a low-entropy source.
func TestRandomString(t *testing.T) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"

	seen := make(map[string]struct{}, 1000)
	for i := 0; i < 1000; i++ {
		s, err := randomString(32)
		if err != nil {
			t.Fatalf("randomString: %v", err)
		}
		if len(s) != 32 {
			t.Fatalf("len = %d; want 32", len(s))
		}
		for _, r := range s {
			if !strings.ContainsRune(charset, r) {
				t.Fatalf("char %q outside charset", r)
			}
		}
		if _, dup := seen[s]; dup {
			t.Fatalf("duplicate token within 1000 samples: %q", s)
		}
		seen[s] = struct{}{}
	}
}

func TestAuthenticateToken_RejectsExpired(t *testing.T) {
	mod := newTestModule(t, nil)
	id := astral.GenerateIdentity()

	expired, err := mod.CreateAccessToken(id, astral.Duration(-time.Hour))
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if _, err := mod.AuthenticateToken(string(expired.Token)); err == nil {
		t.Fatal("expired token authenticated; want error")
	}
}

func TestAuthenticateToken_AcceptsValid(t *testing.T) {
	mod := newTestModule(t, nil)
	id := astral.GenerateIdentity()

	valid, err := mod.CreateAccessToken(id, astral.Duration(time.Hour))
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := mod.AuthenticateToken(string(valid.Token))
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	if !got.IsEqual(id) {
		t.Fatalf("identity = %v; want %v", got, id)
	}
}
