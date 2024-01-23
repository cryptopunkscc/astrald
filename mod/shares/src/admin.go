package shares

import (
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/shares"
	"time"
)

const defaultAccessDuration = time.Hour * 24 * 365 * 100 // 100 years

type Admin struct {
	mod  *Module
	cmds map[string]func(admin.Terminal, []string) error
}

func NewAdmin(mod *Module) *Admin {
	var adm = &Admin{mod: mod}
	adm.cmds = map[string]func(admin.Terminal, []string) error{
		"grant":  adm.grant,
		"revoke": adm.revoke,
		"list":   adm.list,
		"help":   adm.help,
	}

	return adm
}

func (adm *Admin) Exec(term admin.Terminal, args []string) error {
	if len(args) < 2 {
		return adm.help(term, []string{})
	}

	cmd, args := args[1], args[2:]
	if fn, found := adm.cmds[cmd]; found {
		return fn(term, args)
	}

	return errors.New("unknown command")
}

func (adm *Admin) grant(term admin.Terminal, args []string) error {
	if len(args) < 2 {
		return errors.New("argument missing")
	}

	var err error
	var identity id.Identity
	var dataID data.ID

	if identity, err = adm.mod.node.Resolver().Resolve(args[0]); err != nil {
		return err
	}
	if dataID, err = data.Parse(args[1]); err != nil {
		return err
	}

	return adm.mod.Grant(identity, dataID)
}

func (adm *Admin) revoke(term admin.Terminal, args []string) error {
	if len(args) < 2 {
		return errors.New("argument missing")
	}

	var err error
	var identity id.Identity
	var dataID data.ID

	if identity, err = adm.mod.node.Resolver().Resolve(args[0]); err != nil {
		return err
	}
	if dataID, err = data.Parse(args[1]); err != nil {
		return err
	}

	return adm.mod.Revoke(identity, dataID)
}

func (adm *Admin) list(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	identity, err := adm.mod.node.Resolver().Resolve(args[0])
	if err != nil {
		return err
	}

	var indexName = adm.mod.localShareIndexName(identity)

	entries, err := adm.mod.index.UpdatedSince(indexName, time.Time{})

	for _, entry := range entries {
		if entry.Added {
			term.Printf("%v\n", entry.DataID)
		}
	}

	return nil
}

func (adm *Admin) ShortDescription() string {
	return "manage data sharing"
}

func (adm *Admin) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: %s <command>\n\n", shares.ModuleName)
	term.Printf("commands:\n")
	term.Printf("  grant <identity> <dataID> [duration]      grant access to data\n")
	term.Printf("  revoke <identity> <dataID>                revoke access to data\n")
	term.Printf("  list <identitiy>                          list data shared with an identity\n")
	term.Printf("  help                                      show help\n")
	return nil
}
