package nat

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/node/infra/drivers/inet"
)

type natMapping struct {
	intAddr inet.Endpoint
	extAddr inet.Endpoint
}

func (m natMapping) String() string {
	return fmt.Sprintf("%s -> %s", m.intAddr.String(), m.extAddr.String())
}
