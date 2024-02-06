package zip

import "github.com/cryptopunkscc/astrald/data"

const ModuleName = "zip"
const DBPrefix = "zip__"

type Module interface {
	Index(zipID data.ID, reindex bool) error
}
