package zip

import "github.com/cryptopunkscc/astrald/data"

const ArchivesIndexName = "mod.zip.archives"

type Module interface {
	Index(zipID data.ID, reindex bool) error
}
