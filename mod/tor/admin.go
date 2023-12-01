package tor

import (
	"github.com/cryptopunkscc/astrald/mod/admin/api"
)

type Admin struct {
	mod *Module
}

func NewAdmin(mod *Module) *Admin {
	var adm = &Admin{mod: mod}

	return adm
}

func (adm *Admin) Exec(term admin.Terminal, args []string) error {
	return nil
}

func (adm *Admin) ShortDescription() string {
	return "manage the tor driver"
}
