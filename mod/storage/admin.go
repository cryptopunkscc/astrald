package storage

import (
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/admin/api"
	"io"
	"reflect"
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
		"grant":   adm.grant,
		"revoke":  adm.revoke,
		"access":  adm.access,
		"read":    adm.read,
		"add":     adm.add,
		"rm":      adm.rm,
		"ls":      adm.ls,
		"rescan":  adm.rescan,
		"sources": adm.sources,
		"help":    adm.help,
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

func (adm *Admin) access(term admin.Terminal, args []string) error {
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

func (adm *Admin) read(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	for _, idstr := range args {
		dataID, err := data.Parse(idstr)
		if err != nil {
			return err
		}

		r, err := adm.mod.Read(dataID, 0, 0)
		if err != nil {
			return err
		}

		io.Copy(term, r)
	}

	return nil
}

func (adm *Admin) rescan(term admin.Terminal, args []string) error {
	adm.mod.localFiles.Rescan(adm.mod.ctx)

	return nil
}

func (adm *Admin) add(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing arguments")
	}

	return adm.mod.localFiles.AddDir(adm.mod.ctx, args[0])
}

func (adm *Admin) rm(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing arguments")
	}

	return adm.mod.localFiles.RemoveDir(adm.mod.ctx, args[0])
}

func (adm *Admin) ls(term admin.Terminal, args []string) error {
	var prefix = ""

	if len(args) > 0 {
		prefix = args[0]
	}

	var files = adm.mod.localFiles.findByPrefix(prefix)
	for _, file := range files {
		term.Printf("%-68s %v\n", file.ID, file.Path)
	}
	return nil
}

func (adm *Admin) sources(term admin.Terminal, args []string) error {
	term.Printf("Enabled data sources:\n")
	for _, source := range adm.mod.Readers() {
		term.Printf("%s\n", reflect.TypeOf(source))
	}

	return nil
}

func (adm *Admin) ShortDescription() string {
	return "manage storage"
}

func (adm *Admin) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: contacts <command>\n\n")
	term.Printf("commands:\n")
	term.Printf("  grant <identity> <dataID> [duration]      grant access to data\n")
	term.Printf("  revoke <identity> <dataID>                revoke access to data\n")
	term.Printf("  access                                    show all access entries\n")
	term.Printf("  add <dir>                                 add local directory to the index\n")
	term.Printf("  rm <dir>                                  remove local directory from the index\n")
	term.Printf("  ls [prefix]                               list all indexed local files\n")
	term.Printf("  read [dataID]                             read a file by ID (caution - may print binary data)\n")
	term.Printf("  help                                      show help\n")
	return nil
}
