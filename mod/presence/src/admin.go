package presence

import (
	"errors"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/presence"
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
		"list":    adm.list,
		"visible": adm.visible,
		"set":     adm.set,
		"clear":   adm.clear,
		"flags":   adm.flags,
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

func (adm *Admin) visible(term admin.Terminal, args []string) error {
	if len(args) == 0 {
		return errors.New("missing argument")
	}

	var v bool
	switch args[0] {
	case "f", "false", "n", "no", "off":
	case "t", "true", "y", "yes", "on":
		v = true
	default:
		return errors.New("invalid value")
	}

	return adm.mod.SetVisible(v)
}

func (adm *Admin) set(term admin.Terminal, args []string) error {
	if len(args) == 0 {
		return errors.New("missing argument")
	}

	return adm.mod.SetFlags(args[0:]...)
}

func (adm *Admin) clear(term admin.Terminal, args []string) error {
	if len(args) == 0 {
		return errors.New("missing argument")
	}

	return adm.mod.ClearFlags(args[0:]...)
}

func (adm *Admin) flags(term admin.Terminal, args []string) error {
	for _, f := range adm.mod.flags.Clone() {
		term.Printf("%s\n", f)
	}
	return nil
}

func (adm *Admin) list(term admin.Terminal, args []string) error {
	f := "%-20s %-20s %-20s %-12s %s\n"
	term.Printf(f,
		admin.Header("ID"),
		admin.Header("Alias"),
		admin.Header("Endpoint"),
		admin.Header("Age"),
		admin.Header("Flags"),
	)
	for _, ad := range adm.mod.discover.RecentAds() {
		term.Printf(f,
			ad.Identity,
			ad.Alias,
			ad.Endpoint,
			time.Since(ad.Timestamp).Round(time.Second),
			strings.Join(ad.Flags, ","),
		)
	}
	return nil
}

func (adm *Admin) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: %s <command>\n\n", presence.ModuleName)
	term.Printf("commands:\n")
	term.Printf("  list                                      show present identities\n")
	term.Printf("  help                                      show help\n")
	return nil
}

func (adm *Admin) ShortDescription() string {
	return "list present identities"
}
