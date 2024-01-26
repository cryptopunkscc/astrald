package shares

import (
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/shares"
	"strings"
	"time"
)

type Admin struct {
	mod  *Module
	cmds map[string]func(admin.Terminal, []string) error
}

func NewAdmin(mod *Module) *Admin {
	var adm = &Admin{mod: mod}
	adm.cmds = map[string]func(admin.Terminal, []string) error{
		"grant":   adm.grant,
		"revoke":  adm.revoke,
		"local":   adm.local,
		"remote":  adm.remote,
		"sync":    adm.sync,
		"syncall": adm.syncAll,
		"unsync":  adm.unsync,
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

func (adm *Admin) syncAll(term admin.Terminal, args []string) error {
	var rows []dbRemoteShare

	var tx = adm.mod.db.Find(&rows)
	if tx.Error != nil {
		return tx.Error
	}

	for _, row := range rows {
		term.Printf("syncing %v@%v... ", row.Caller, row.Target)

		err := adm.mod.Sync(row.Caller, row.Target)
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

	err = adm.mod.Sync(caller, target)
	if err == nil {
		term.Printf("synced %v@%v\n", caller, target)
	}

	return err
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

func (adm *Admin) unsync(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	caller, target, err := adm.parseCallerAndTarget(args[0], term.UserIdentity())
	if err != nil {
		return err
	}

	err = adm.mod.Unsync(caller, target)
	if err == nil {
		term.Printf("unsynced %v@%v\n", caller, target)
	}

	return err
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
	if dataID, err = data.Parse(args[1]); err == nil {
		return adm.mod.Grant(identity, dataID)
	}
	if _, err = adm.mod.index.IndexInfo(args[1]); err == nil {
		return adm.mod.GrantIndex(identity, args[1])
	}

	return errors.New("invalid target")
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
	if dataID, err = data.Parse(args[1]); err == nil {
		return adm.mod.Revoke(identity, dataID)

	}
	if _, err = adm.mod.index.IndexInfo(args[1]); err == nil {
		return adm.mod.RevokeIndex(identity, args[1])
	}

	return errors.New("invalid target")
}

func (adm *Admin) local(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	identity, err := adm.mod.node.Resolver().Resolve(args[0])
	if err != nil {
		return err
	}

	var indexName = adm.mod.localShareIndexName(identity)

	entries, err := adm.mod.index.UpdatedBetween(indexName, time.Time{}, time.Time{})

	for _, entry := range entries {
		if entry.Added {
			term.Printf("%v\n", entry.DataID)
		}
	}

	return nil
}

func (adm *Admin) remote(term admin.Terminal, args []string) error {
	if len(args) == 0 {
		shares, err := adm.mod.RemoteShares()
		if err != nil {
			return err
		}

		term.Printf("Synced shares:\n")
		for _, share := range shares {
			term.Printf("%v@%v\n", share.Caller, share.Target)
		}
	} else {
		caller, target, err := adm.parseCallerAndTarget(args[0], term.UserIdentity())
		if err != nil {
			return err
		}

		list, err := adm.mod.ListRemote(caller, target)
		if err != nil {
			return err
		}

		for _, item := range list {
			term.Printf("%v\n", item)
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
	term.Printf("  grant <identity> <dataID|index>           grant access to data or an index\n")
	term.Printf("  revoke <identity> <dataID|index>          revoke access to data or an index\n")
	term.Printf("  local <identitiy>                         list local shres for the identity\n")
	term.Printf("  remote [identity]                         list remote shares\n")
	term.Printf("  sync [guest@]<host>                       sync remote share\n")
	term.Printf("  unsync [guest@]<host>                     unsync remote share (remove and stop following)\n")
	term.Printf("  syncall                                   sync all remote shares\n")
	term.Printf("  help                                      show help\n")
	return nil
}
