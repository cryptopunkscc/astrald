package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"slices"
)

type opListTokensArgs struct {
	ID  *astral.Identity `query:"optional"`
	Out string           `query:"optional"`
}

func (mod *Module) OpListTokens(ctx *astral.Context, q shell.Query, args opListTokensArgs) (err error) {
	tokens, err := mod.ListAccessTokens()
	if err != nil {
		mod.log.Errorv(1, "ListAccessTokens: %v", err)
		return q.RejectWithCode(astral.CodeInternalError)
	}

	tokens = slices.DeleteFunc(tokens, func(token *apphost.AccessToken) bool {
		return token.Identity.IsEqual(args.ID)
	})

	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	for _, token := range tokens {
		err = ch.Write(token)
		if err != nil {
			return err
		}
	}

	return
}
