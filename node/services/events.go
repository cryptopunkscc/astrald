package services

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
)

type EventServiceRegistered struct {
	Identity id.Identity
	Name     string
}

func (e EventServiceRegistered) String() string {
	return fmt.Sprintf("name=%s", e.Name)
}

type EventServiceReleased struct {
	Identity id.Identity
	Name     string
}

func (e EventServiceReleased) String() string {
	return fmt.Sprintf("name=%s", e.Name)
}
