package apphost

import (
	"slices"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opListTokensArgs struct {
	ID  *astral.Identity `query:"optional"`
	Out string           `query:"optional"`
}

// OpListTokens lists all access tokens of an identity
func (mod *Module) OpListTokens(ctx *astral.Context, q shell.Query, args opListTokensArgs) (err error) {
	ch := channel.New(q.Accept(), channel.OutFmt(args.Out))
	defer ch.Close()

	// get token list
	tokens, err := mod.ListAccessTokens()
	if err != nil {
		ch.Write(astral.NewError("internal error"))
		return err
	}

	// filter tokens by ID
	if !args.ID.IsZero() {
		tokens = slices.DeleteFunc(tokens, func(token *apphost.AccessToken) bool {
			return !token.Identity.IsEqual(args.ID)
		})
	}

	for _, token := range tokens {
		err = ch.Write(token)
		if err != nil {
			return err
		}
	}

	return ch.Write(&astral.EOS{})
}
