package fwd

import (
	"errors"
	"github.com/cryptopunkscc/astrald/mod/admin"
)

type Admin struct {
	mod  *Module
	cmds map[string]func(admin.Terminal, []string) error
}

func NewAdmin(mod *Module) *Admin {
	var adm = &Admin{mod: mod}
	adm.cmds = map[string]func(admin.Terminal, []string) error{
		"list":  adm.list,
		"start": adm.start,
		"stop":  adm.stop,
		"help":  adm.help,
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

func (adm *Admin) list(term admin.Terminal, args []string) error {
	var f = "%-39s %-39s\n"
	term.Printf(f, admin.Header("Server"), admin.Header("Target"))
	for _, server := range adm.mod.Servers() {
		term.Printf(
			f,
			server.Server,
			server.Server.Target(),
		)
	}

	return nil
}

func (adm *Admin) stop(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	for _, server := range adm.mod.Servers() {
		if server.String() == args[0] {
			term.Printf("stopping %v... ", server)
			server.Stop()
			<-server.Done()
			term.Printf("ok\n")
			return nil
		}
	}

	return errors.New("server not found")
}

func (adm *Admin) start(term admin.Terminal, args []string) error {
	if len(args) < 2 {
		return errors.New("missing argument")
	}

	term.Printf("creating forward... ")

	err := adm.mod.CreateForward(args[0], args[1])

	if err != nil {
		term.Printf("%v\n", err)
	} else {
		term.Printf("ok\n")
	}

	return nil
}

func (adm *Admin) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: %s <command>\n\n", ModuleName)
	term.Printf("commands:\n")
	var f = "  %-26s %s\n"
	term.Printf(f, "list", "list running servers")
	term.Printf(f, "start <server> <target>", "start a new forward")
	term.Printf(f, "stop <server>", "stop a forward")
	term.Printf(f, "help", "show help")
	return nil
}

func (adm *Admin) ShortDescription() string {
	return "cross-network forwarding tool"
}
