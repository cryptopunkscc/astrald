package user

import (
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/auth"
)

// IsNodeContract reports whether a contract grants SwarmAccess.
func IsNodeContract(c *auth.Contract) bool {
	return len(c.HasPermit(SwarmAccessAction{}.ObjectType())) > 0
}

// NewNodeContract creates a node contract granting SwarmAccess from issuer to subject.
func NewNodeContract(issuer, subject *astral.Identity, duration time.Duration) (*auth.Contract, error) {
	permits := []*auth.Permit{
		{Action: astral.String8(SwarmAccessAction{}.ObjectType())},
	}

	return &auth.Contract{
		Issuer:    issuer,
		Subject:   subject,
		Permits:   astral.WrapSlice(&permits),
		ExpiresAt: astral.Time(time.Now().Add(duration)),
	}, nil
}
