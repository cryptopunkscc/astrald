package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"strings"
	"time"
)

func (mod *Module) OpSearch(ctx *astral.Context, q shell.Query, args objects.SearchArgs) (err error) {
	opts := objects.DefaultSearchOpts()
	opts.ClientID = q.Caller()

	// handle args for local queries
	if q.Origin() == "" {
		ctx = ctx.WithZone(args.Zone)

		if len(args.Ext) > 0 {
			var ids []*astral.Identity
			targets := strings.Split(args.Ext, ",")
			for _, target := range targets {
				id, err := mod.Dir.ResolveIdentity(target)
				if err != nil {
					return q.Reject()
				}
				ids = append(ids, id)
			}
			opts.Extra.Set("ext", ids)
		}
	}

	sctx, cancel := ctx.WithTimeout(time.Minute)
	defer cancel()

	matches, err := mod.Search(sctx, args.Query, opts)
	if err != nil {
		cancel()
		return q.Reject()
	}

	ch := astral.NewChannel(q.Accept(), args.Format)
	defer ch.Close()

	var dup = make(map[string]struct{})

	for match := range matches {
		if !mod.Auth.Authorize(q.Caller(), objects.ActionRead, &match.ObjectID) {
			continue
		}

		if _, found := dup[match.ObjectID.String()]; found {
			continue
		}

		dup[match.ObjectID.String()] = struct{}{}

		err = ch.Write(match)
		if err != nil {
			return
		}
	}

	return nil
}
