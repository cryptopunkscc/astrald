package apphost

import (
	"crypto/rand"
	"errors"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/sig"
)

func (mod *Module) ListAccessTokens() ([]*apphost.AccessToken, error) {
	rows, err := mod.db.ListAccessTokens()
	if err != nil {
		return nil, err
	}

	return sig.MapSlice(rows, func(a dbAccessToken) (*apphost.AccessToken, error) {
		return &apphost.AccessToken{
			Identity:  a.Identity,
			Token:     astral.String8(a.Token),
			ExpiresAt: astral.Time(a.ExpiresAt),
		}, nil
	})
}

func (mod *Module) CreateAccessToken(identity *astral.Identity, d astral.Duration) (*apphost.AccessToken, error) {
	token, err := mod.db.CreateAccessToken(identity, d)
	if err != nil {
		return nil, err
	}

	return &apphost.AccessToken{
		Identity:  token.Identity,
		Token:     astral.String8(token.Token),
		ExpiresAt: astral.Time(token.ExpiresAt),
	}, nil
}

// AuthenticateToken resolves a bearer token to the identity it was issued for.
// Any lookup or expiry failure is collapsed into a single opaque error to avoid leaking token existence.
func (mod *Module) AuthenticateToken(token string) (*astral.Identity, error) {
	dbToken, err := mod.db.FindAccessToken(token)
	if err != nil || dbToken == nil {
		return nil, errors.New("invalid token")
	}

	// reject expired tokens; collapsed into the same opaque error
	if time.Now().After(dbToken.ExpiresAt) {
		return nil, errors.New("invalid token")
	}

	return dbToken.Identity, nil
}

// randomString returns a cryptographically random string of length characters
// over [a-zA-Z0-9_].
// why: access tokens are bearer credentials; math/rand is predictable from
// observed output and must never generate them.
func randomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	// reject the biased tail so every charset index is equally likely
	const maxByte = 256 - (256 % len(charset))

	out := make([]byte, length)
	buf := make([]byte, length)
	for filled := 0; filled < length; {
		if _, err := rand.Read(buf); err != nil {
			return "", err
		}
		for _, b := range buf {
			if filled == length {
				break
			}
			if int(b) < maxByte {
				out[filled] = charset[int(b)%len(charset)]
				filled++
			}
		}
	}
	return string(out), nil
}
