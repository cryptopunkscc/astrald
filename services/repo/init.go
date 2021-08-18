package repo

import (
	"github.com/cryptopunkscc/astrald/node"
)

func init() {

	// Start local read write repository
	_ = node.RegisterService(Port, NewRepoService().Run)

	// Start remote read only repository
	_ = node.RegisterService(FilesPort, NewFilesService().Run)
}
