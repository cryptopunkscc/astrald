package nat

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/infra/inet"
)

type natMapping struct {
	intAddr inet.Addr
	extAddr inet.Addr
}

func (m natMapping) String() string {
	return fmt.Sprintf("%s -> %s", m.intAddr.String(), m.extAddr.String())
}
