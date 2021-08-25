package init

import (
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/services/repo"
)

func init() {

	// Start local read write repository
	_ = node.RegisterService(repo.Port, repo.NewRepoService().Run)

	// Start remote read only repository
	_ = node.RegisterService(repo.FilesPort, repo.NewFilesService().Run)
}
