package main

import (
	"github.com/cryptopunkscc/astrald/astral"
	dircli "github.com/cryptopunkscc/astrald/mod/dir/client"
	modlog "github.com/cryptopunkscc/astrald/mod/log"
)

func loadAliasMap(ctx *astral.Context) (err error) {
	aliasMap, err := dircli.Default().AliasMap(ctx)
	if err != nil {
		return
	}

	modlog.IdentityResolver.Set(newResolver(aliasMap))

	return
}
