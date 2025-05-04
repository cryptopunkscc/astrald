package objects

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"time"
)

func (mod *Module) OpSearch(ctx *astral.Context, q shell.Query, args objects.SearchArgs) (err error) {
	ctx, cancel := ctx.WithIdentity(q.Caller()).IncludeZone(args.Zone).WithTimeout(time.Minute)
	defer cancel()

	opts := objects.DefaultSearchOpts()
	opts.ClientID = q.Caller()

	matches, err := mod.Search(ctx, args.Query, opts)
	if err != nil {
		return q.Reject()
	}

	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	var dup = make(map[string]struct{})

	for match := range matches {
		if !mod.Auth.Authorize(q.Caller(), objects.ActionRead, match.ObjectID) {
			continue
		}

		if _, found := dup[match.ObjectID.String()]; found {
			continue
		}

		if args.Access {
			if has, _ := mod.root.Contains(ctx, match.ObjectID); !has {
				continue
			}
		}

		dup[match.ObjectID.String()] = struct{}{}

		err = ch.Write(match)
		if err != nil {
			return fmt.Errorf("error writing match: %w", err)
		}
	}

	return nil
}
