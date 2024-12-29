package tcp

import (
	"errors"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/tcp"
)

type Admin struct {
	mod  *Module
	cmds map[string]func(admin.Terminal, []string) error
}

func NewAdmin(mod *Module) *Admin {
	var adm = &Admin{mod: mod}
	adm.cmds = map[string]func(admin.Terminal, []string) error{
		"info": adm.info,
		"help": adm.help,
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

func (adm *Admin) info(term admin.Terminal, _ []string) error {
	term.Printf("Listen port: %d\n\n", adm.mod.ListenPort())

	ips, _ := adm.mod.localIPs()
	if len(ips) == 0 {
		term.Printf("No local IP addresses detected.\n\n")
	} else {
		term.Printf("Local IP addresses:\n")
		for _, ip := range ips {
			term.Printf("- %v\n", ip)
		}
		term.Printf("\n")
	}

	endpoints := adm.mod.localEndpoints()
	if len(endpoints) == 0 {
		term.Printf("No local endpoints.\n\n")
	} else {
		term.Printf("Local endpoints:\n")
		for _, e := range endpoints {
			term.Printf("- %v\n", e)
		}
		term.Printf("\n")
	}

	if len(adm.mod.publicEndpoints) == 0 {
		term.Printf("No custom endpoints.\n")
	} else {
		term.Printf("Custom endpoints:\n")
		for _, endpoint := range adm.mod.publicEndpoints {
			term.Printf("- %v\n", endpoint)
		}
		term.Printf("\n")
	}

	return nil
}

func (adm *Admin) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: %v <command>\n\n", tcp.ModuleName)
	term.Printf("commands:\n")
	term.Printf("  info            show tcp driver info\n")
	term.Printf("  help            show help\n")
	return nil
}

func (adm *Admin) ShortDescription() string {
	return "manage the tcp driver"
}
