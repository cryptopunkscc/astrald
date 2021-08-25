package init

import (
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/services/scanner"
)

func init() {
	_ = node.RegisterService(scanner.Port, scanner.Run)
}
