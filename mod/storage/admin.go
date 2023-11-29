package storage

import (
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/admin/api"
	"time"
)

const defaultAccessDuration = time.Hour * 24 * 365 * 100

type Admin struct {
	mod  *Module
	cmds map[string]func(admin.Terminal, []string) error
}

func NewAdmin(mod *Module) *Admin {
	var adm = &Admin{mod: mod}
	adm.cmds = map[string]func(admin.Terminal, []string) error{
		"grant":           adm.grant,
		"revoke":          adm.revoke,
		"list":            adm.list,
		"sources":         adm.sources,
		"providers":       adm.providers,
		"add_provider":    adm.addProvider,
		"remove_provider": adm.removeProvider,
		"help":            adm.help,
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
	var expiresAt = time.Now().Add(defaultAccessDuration)

	if identity, err = adm.mod.node.Resolver().Resolve(args[0]); err != nil {
		return err
	}
	if dataID, err = data.Parse(args[1]); err != nil {
		return err
	}
	if len(args) >= 3 {
		d, err := time.ParseDuration(args[2])
		if err != nil {
			return err
		}
		expiresAt = time.Now().Add(d)
	}

	return adm.mod.GrantAccess(identity, dataID, expiresAt)
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

	return adm.mod.RevokeAccess(identity, dataID)
}

func (adm *Admin) list(term admin.Terminal, args []string) error {
	var list []dbAccess

	tx := adm.mod.db.Limit(100).Find(&list)
	if tx.Error != nil {
		return tx.Error
	}
	f := "%-20s%-16s%s\n"

	term.Printf("showing %d results\n", len(list))
	term.Printf(f, admin.Header("IDENTITY"), admin.Header("EXPIRY"), admin.Header("DATAID"))

	for _, item := range list {
		access, err := item.toAccess()
		if err != nil {
			term.Printf("data error: %s\n", err)
			continue
		}

		var expiry = "expired"

		if access.ExpiresAt.After(time.Now()) {
			expiry = time.Until(access.ExpiresAt).Round(time.Second).String()
		}

		term.Printf(f,
			access.Identity,
			expiry,
			access.DataID)
	}

	return nil
}

func (adm *Admin) sources(term admin.Terminal, args []string) error {
	var sources = adm.mod.DataSources()

	term.Printf("%d registered data source(s)\n", len(sources))
	var fmt = "%-20s %s\n"
	term.Printf(fmt, admin.Header("IDENTITY"), admin.Header("SERVICE"))
	for _, source := range sources {
		term.Printf(fmt, source.Identity, source.Service)
	}

	return nil
}

func (adm *Admin) providers(term admin.Terminal, args []string) error {
	list, err := adm.mod.AllProviders()
	if err != nil {
		return err
	}

	for _, identity := range list {
		term.Printf("%s\n", identity)
	}

	return nil
}

func (adm *Admin) addProvider(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("usage: addprovider <identity>")
	}

	identity, err := adm.mod.node.Resolver().Resolve(args[0])
	if err != nil {
		return err
	}

	return adm.mod.AddProvider(identity)
}

func (adm *Admin) removeProvider(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("usage: remove_provider <identity>")
	}

	identity, err := adm.mod.node.Resolver().Resolve(args[0])
	if err != nil {
		return err
	}

	return adm.mod.RemoveProvider(identity)
}

func (adm *Admin) ShortDescription() string {
	return "manage storage providers and data access"
}

func (adm *Admin) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: contacts <command>\n\n")
	term.Printf("commands:\n")
	term.Printf("  grant <identity> <dataID> [duration]      grant access to data\n")
	term.Printf("  revoke <identity> <dataID>                grant access to data\n")
	term.Printf("  list                                      list access entries\n")
	term.Printf("  sources                                   list registered sources\n")
	term.Printf("  providers                                 list identities allowed as providers\n")
	term.Printf("  add_provider <identity>                   add a provider\n")
	term.Printf("  remove_provider <identity>                remove a provider\n")
	term.Printf("  help                                      show help\n")
	return nil
}
