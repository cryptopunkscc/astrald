package apphost

import (
	"encoding/json"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"time"
)

const DefaultTokenDuration = astral.Duration(time.Hour * 24 * 365) // 1 year

type opCreateTokenArgs struct {
	ID       *astral.Identity
	Duration astral.Duration `query:"optional"`
	Format   astral.String   `query:"optional"`
}

func (mod *Module) OpCreateToken(ctx *astral.Context, q shell.Query, args opCreateTokenArgs) (err error) {
	if args.ID.IsZero() {
		return q.Reject()
	}

	if args.Duration == 0 {
		args.Duration = DefaultTokenDuration
	}

	mod.log.Infov(1, "creating token for %v valid for %v", args.ID, args.Duration)

	token, err := mod.CreateAccessToken(args.ID, args.Duration)
	if err != nil {
		mod.log.Errorv(1, "error creating token for %v: %v", args.ID, err)
		return q.Reject()
	}

	conn := q.Accept()
	defer conn.Close()

	switch args.Format {
	case "json":
		err = json.NewEncoder(conn).Encode(token)
	case "bin", "":
		_, err = token.WriteTo(conn)
	default:
		return errors.New("unsupported format")
	}

	return
}
