package main

import (
	"github.com/cryptopunkscc/astrald/lib/astrald"
	modlog "github.com/cryptopunkscc/astrald/mod/log"
)

func loadAliasMap() (err error) {
	aliasMap, err := astrald.Dir().AliasMap()
	if err != nil {
		return
	}

	modlog.IdentityResolver.Set(newResolver(aliasMap))

	return
}
