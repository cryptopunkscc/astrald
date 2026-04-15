package objects

import (
	"fmt"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/routing"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

type SearchArgs struct {
	Query string      `query:"key:q"`
	Repo  string      `query:"optional"` // return only objects that this repo contains
	Zone  astral.Zone `query:"optional"`
	Out   string      `query:"optional"`
}

func (mod *Module) OpSearch(ctx *astral.Context, q *routing.IncomingQuery, args SearchArgs) (err error) {
	ctx, cancel := ctx.WithIdentity(q.Caller()).IncludeZone(args.Zone).WithTimeout(time.Minute)
	defer cancel()

	ch := q.Accept(channel.WithOutputFormat(args.Out))
	defer ch.Close()

	// find repo if provided
	var repo objects.Repository
	if len(args.Repo) > 0 {
		repo = mod.GetRepository(args.Repo)
		if repo == nil {
			return ch.Send(astral.NewError("repository not found"))
		}
	}

	var searchQuery objects.SearchQuery
	_ = searchQuery.UnmarshalText([]byte(args.Query))

	// run the search
	matches, err := mod.Search(ctx, searchQuery)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	var dup = make(map[string]struct{})

	for match := range matches {
		// deduplicate results
		if _, found := dup[match.ObjectID.String()]; found {
			continue
		}

		// filter by repository if provided
		if repo != nil {
			contains, err := repo.Contains(ctx, match.ObjectID)
			switch {
			case err != nil:
				continue
			case !contains:
				continue
			}
		}

		dup[match.ObjectID.String()] = struct{}{}

		err = ch.Send(match)
		if err != nil {
			return fmt.Errorf("error writing match: %w", err)
		}
	}

	return nil
}
