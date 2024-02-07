package shares

import (
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/shares"
	"strings"
)

type Admin struct {
	mod  *Module
	cmds map[string]func(admin.Terminal, []string) error
}

func NewAdmin(mod *Module) *Admin {
	var adm = &Admin{mod: mod}
	adm.cmds = map[string]func(admin.Terminal, []string) error{
		"add":     adm.add,
		"remove":  adm.remove,
		"local":   adm.local,
		"remote":  adm.remote,
		"sync":    adm.sync,
		"syncall": adm.syncAll,
		"unsync":  adm.unsync,
		"notify":  adm.notify,
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

func (adm *Admin) add(term admin.Terminal, args []string) error {
	if len(args) < 2 {
		return errors.New("argument missing")
	}

	var err error
	var identity id.Identity
	var dataID data.ID

	if identity, err = adm.mod.node.Resolver().Resolve(args[0]); err != nil {
		return err
	}

	share, err := adm.mod.LocalShare(identity, true)
	if err != nil {
		return err
	}

	if dataID, err = data.Parse(args[1]); err == nil {
		return share.AddObject(dataID)
	}

	return share.AddSet(args[1])
}

func (adm *Admin) remove(term admin.Terminal, args []string) error {
	if len(args) < 2 {
		return errors.New("argument missing")
	}

	var err error
	var identity id.Identity
	var dataID data.ID

	if identity, err = adm.mod.node.Resolver().Resolve(args[0]); err != nil {
		return err
	}

	share, err := adm.mod.LocalShare(identity, false)
	if err != nil {
		return err
	}

	if dataID, err = data.Parse(args[1]); err == nil {
		return share.RemoveObject(dataID)
	}

	return share.RemoveSet(args[1])
}

func (adm *Admin) syncAll(term admin.Terminal, args []string) error {
	var rows []dbRemoteShare

	var tx = adm.mod.db.Find(&rows)
	if tx.Error != nil {
		return tx.Error
	}

	for _, row := range rows {
		term.Printf("syncing %v@%v... ", row.Caller, row.Target)

		share, err := adm.mod.findRemoteShare(row.Caller, row.Target)
		if err != nil {
			term.Printf("%v\n", err)
			continue
		}

		err = share.Sync()
		if err != nil {
			term.Printf("%v\n", err)
			continue
		}

		term.Printf("ok\n")
	}

	return nil
}

func (adm *Admin) sync(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	caller, target, err := adm.parseCallerAndTarget(args[0], term.UserIdentity())
	if err != nil {
		return err
	}

	share, err := adm.mod.FindOrCreateRemoteShare(caller, target)
	if err != nil {
		return err
	}

	err = share.Sync()
	if err == nil {
		term.Printf("synced %v@%v\n", caller, target)
	}

	return err
}

func (adm *Admin) unsync(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	caller, target, err := adm.parseCallerAndTarget(args[0], term.UserIdentity())
	if err != nil {
		return err
	}

	share, err := adm.mod.FindRemoteShare(caller, target)
	if err != nil {
		return err
	}

	err = share.Unsync()
	if err == nil {
		term.Printf("unsynced %v@%v\n", caller, target)
	}

	return err
}

func (adm *Admin) notify(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	identity, err := adm.mod.node.Resolver().Resolve(args[0])
	if err != nil {
		return err
	}

	return adm.mod.Notify(identity)
}

func (adm *Admin) local(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	identity, err := adm.mod.node.Resolver().Resolve(args[0])
	if err != nil {
		return err
	}

	share, err := adm.mod.FindLocalShare(identity)
	if err != nil {
		return err
	}

	entries, err := share.Scan(nil)
	for _, entry := range entries {
		term.Printf("%v\n", entry.DataID)
	}

	return nil
}

func (adm *Admin) remote(term admin.Terminal, args []string) error {
	if len(args) == 0 {
		var rows []dbRemoteShare

		var err = adm.mod.db.Find(&rows).Error
		if err != nil {
			return err
		}

		term.Printf("Synced shares:\n")
		for _, row := range rows {
			term.Printf("%v@%v\n", row.Caller, row.Target)
		}
	} else {
		caller, target, err := adm.parseCallerAndTarget(args[0], term.UserIdentity())
		if err != nil {
			return err
		}

		share, err := adm.mod.findRemoteShare(caller, target)
		if err != nil {
			return err
		}

		scan, err := share.set.Scan(nil)
		if err != nil {
			return err
		}

		for _, item := range scan {
			term.Printf("%v\n", item.DataID)
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
	term.Printf("  add <identity> <dataID|set>               add access to data or a set\n")
	term.Printf("  remove <identity> <dataID|set>            remove access to data or a set\n")
	term.Printf("  local <identitiy>                         list local shres for the identity\n")
	term.Printf("  remote [identity]                         list remote shares\n")
	term.Printf("  sync [guest@]<host>                       sync remote share\n")
	term.Printf("  unsync [guest@]<host>                     unsync remote share (remove and stop following)\n")
	term.Printf("  syncall                                   sync all remote shares\n")
	term.Printf("  help                                      show help\n")
	return nil
}

func (adm *Admin) parseCallerAndTarget(targetName string, defaultCaller id.Identity) (
	caller id.Identity,
	target id.Identity,
	err error,
) {
	if i := strings.IndexByte(targetName, '@'); i != -1 {
		caller, err = adm.mod.node.Resolver().Resolve(targetName[:i])
		if err != nil {
			return
		}
		targetName = targetName[i+1:]
	} else {
		caller = defaultCaller
	}

	target, err = adm.mod.node.Resolver().Resolve(targetName)

	return
}
