package shares

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/lib/arl"
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
		"remote":     adm.remote,
		"sync":       adm.sync,
		"rawsync":    adm.rawsync,
		"syncall":    adm.syncAll,
		"unsync":     adm.unsync,
		"purgecache": adm.purgecache,
		"help":       adm.help,
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

		share, err := adm.mod.findRemoteShare(row.Caller, row.Target)
		if err != nil {
			term.Printf("%v\n", err)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		err = share.Sync(ctx)
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

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err = share.Sync(ctx)
	if err == nil {
		term.Printf("synced %v@%v\n", caller, target)
	}

	return err
}

func (adm *Admin) rawsync(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	a, err := arl.Parse(args[0], adm.mod.Dir)
	if err != nil {
		return err
	}

	if a.Caller.IsZero() {
		a.Caller = term.UserIdentity()
	}

	if a.Query == "" {
		a.Query = "shares.sync"
	}

	c := NewConsumer(adm.mod, a.Caller, a.Target)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	diff, err := c.Sync(ctx, time.Time{})
	if err != nil {
		return err
	}

	term.Printf("%-64s %v\n", admin.Header("ObjectID"), admin.Header("Present"))
	for _, u := range diff.Updates {
		term.Printf("%-64s %v\n", u.ObjectID, u.Present)
	}

	term.Printf("%v\n", diff.Time)

	return nil
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

func (adm *Admin) purgecache(term admin.Terminal, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: purgecache <minAge>")
	}

	cache := &DescriptorCache{mod: adm.mod}
	minAge, err := time.ParseDuration(args[0])
	if err != nil {
		return err
	}

	return cache.Purge(minAge)
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
			term.Printf("%v\n", item.ObjectID)
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
		caller, err = adm.mod.Dir.Resolve(targetName[:i])
		if err != nil {
			return
		}
		targetName = targetName[i+1:]
	} else {
		caller = defaultCaller
	}

	target, err = adm.mod.Dir.Resolve(targetName)

	return
}
