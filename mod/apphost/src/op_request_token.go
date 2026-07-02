package apphost

import (
	"slices"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/routing"
)

// TrustedWebOrigin is the built-in first-party origin, seeded into the default
// TrustedWebOrigins allowlist. Operators add further origins (e.g. a dev server)
// via config.
const TrustedWebOrigin = "https://settings.astrald.app"

// originAllowed reports whether an anonymous browser guest from origin may route
// queries: origin must be in the configured TrustedWebOrigins allowlist. Callers
// must pass a non-empty origin; an empty origin denotes a non-browser client and
// is governed by AllowAnonymous, not this check.
func (mod *Module) originAllowed(origin string) bool {
	return slices.Contains(mod.config.TrustedWebOrigins, origin)
}

type opRequestTokenArgs struct {
	Out string `query:"optional"`
}

// OpRequestToken mints an access token for the node's active user when the query
// arrives from the trusted web origin. Reject codes: 1 - untrusted or missing
// origin; 2 - node has no active user (the caller should run first-time setup).
func (mod *Module) OpRequestToken(ctx *astral.Context, q *routing.IncomingQuery, args opRequestTokenArgs) (err error) {
	// note: the origin check must run before Accept/Reject - the en-route entry
	// is removed once the query resolves
	o, _ := mod.EnRouteQueryExtra(q.Nonce(), "origin-web")
	origin, _ := o.(string)
	if !mod.originAllowed(origin) {
		return q.Reject()
	}

	// why: the token authenticates as the active user; with no user there is
	// nothing to authenticate as - code 2 mirrors user.info so the caller can
	// branch into first-time setup
	if mod.User == nil || mod.User.Identity().IsZero() {
		return q.RejectWithCode(2)
	}

	token, err := mod.CreateAccessToken(mod.User.Identity(), DefaultTokenDuration)
	if err != nil {
		mod.log.Errorv(1, "error creating token for %v: %v", mod.User.Identity(), err)
		return q.RejectWithCode(astral.CodeInternalError)
	}

	mod.log.Logv(1, "minted access token for %v requested by %v", token.Identity, origin)

	ch := channel.New(q.AcceptRaw(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	return ch.Send(token)
}
