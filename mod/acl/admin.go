package acl

import (
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	acl "github.com/cryptopunkscc/astrald/mod/acl/api"
	"github.com/cryptopunkscc/astrald/mod/admin/api"
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

	return adm.mod.Grant(identity, dataID, expiresAt)
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
	var list []dbPerm

	tx := adm.mod.db.Limit(1000).Find(&list)
	if tx.Error != nil {
		return tx.Error
	}
	f := "%-20s%-16s%s\n"

	term.Printf("showing %d results\n", len(list))
	term.Printf(f, admin.Header("IDENTITY"), admin.Header("EXPIRY"), admin.Header("DATAID"))

	for _, perm := range list {
		userID, err := id.ParsePublicKeyHex(perm.Identity)
		if err != nil {
			continue
		}

		dataID, err := data.Parse(perm.DataID)
		if err != nil {
			continue
		}

		var expiry = "expired"
		if perm.ExpiresAt.After(time.Now()) {
			expiry = time.Until(perm.ExpiresAt).Round(time.Second).String()
		}

		term.Printf(f, userID, expiry, dataID)
	}

	return nil
}

func (adm *Admin) ShortDescription() string {
	return "data access control list"
}

func (adm *Admin) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: %s <command>\n\n", acl.ModuleName)
	term.Printf("commands:\n")
	term.Printf("  grant <identity> <dataID> [duration]      grant access to data\n")
	term.Printf("  revoke <identity> <dataID>                revoke access to data\n")
	term.Printf("  list                                      show all access entries\n")
	term.Printf("  help                                      show help\n")
	return nil
}
