package zip

import "github.com/cryptopunkscc/astrald/data"

const AllArchivedSet = "mod.zip.all_archived"
const ModuleName = "zip"
const DBPrefix = "zip__"

type Module interface {
	Index(zipID data.ID, reindex bool) error
}
