package presence

import (
	"github.com/cryptopunkscc/astrald/mod/admin/api"
	"time"
)

type Admin struct {
	mod *Module
}

func NewAdmin(mod *Module) *Admin {
	var adm = &Admin{mod: mod}

	return adm
}

func (adm *Admin) Exec(term admin.Terminal, args []string) error {
	f := "%-20s %-20s %-12s %-8s %s\n"
	term.Printf(f,
		admin.Header("Alias"),
		admin.Header("Endpoint"),
		admin.Header("Age"),
		admin.Header("Flags"),
		admin.Header("ID"),
	)
	for _, ad := range adm.mod.Discover.RecentAds() {
		var flags = ""
		if ad.DiscoverFlag() {
			flags = flags + "D"
		}
		term.Printf(f,
			ad.Alias,
			ad.Endpoint,
			time.Since(ad.Timestamp).Round(time.Second),
			flags,
			ad.Identity.String(),
		)
	}
	return nil
}

func (adm *Admin) ShortDescription() string {
	return "list present identities"
}
