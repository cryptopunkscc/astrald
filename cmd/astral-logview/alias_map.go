package main

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	modlog "github.com/cryptopunkscc/astrald/mod/log"
)

func loadAliasMap(ctx *astral.Context) (err error) {
	aliasMap, err := astrald.Dir().AliasMap(ctx)
	if err != nil {
		return
	}

	modlog.IdentityResolver.Set(newResolver(aliasMap))

	return
}
