package zip

import "github.com/cryptopunkscc/astrald/object"

const ModuleName = "zip"
const DBPrefix = "zip__"

type Module interface {
	Index(zipID object.ID, reindex bool) error
}
