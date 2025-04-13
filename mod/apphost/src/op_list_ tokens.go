package apphost

import (
	"encoding/json"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"slices"
)

type opListTokensArgs struct {
	ID     *astral.Identity `query:"optional"`
	Format string           `query:"optional"`
}

func (mod *Module) OpListTokens(ctx *astral.Context, q shell.Query, args opListTokensArgs) (err error) {
	tokens, err := mod.ListAccessTokens()
	if err != nil {
		mod.log.Errorv(1, "ListAccessTokens: %v", err)
		return q.Reject()
	}

	tokens = slices.DeleteFunc(tokens, func(token *apphost.AccessToken) bool {
		return token.Identity.IsEqual(args.ID)
	})

	conn := q.Accept()
	defer conn.Close()

	switch args.Format {
	case "json":
		err = json.NewEncoder(conn).Encode(tokens)
	case "bin", "":
		_, err = astral.WrapSlice(&tokens, 32, 16).WriteTo(conn)
	default:
		return errors.New("unsupported format")
	}

	return
}
