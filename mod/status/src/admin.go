package status

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/status"
)

type Admin struct {
	mod  *Module
	cmds map[string]func(admin.Terminal, []string) error
}

func NewAdmin(mod *Module) *Admin {
	var adm = &Admin{mod: mod}
	adm.cmds = map[string]func(admin.Terminal, []string) error{
		"scan":    adm.scan,
		"show":    adm.show,
		"update":  adm.update,
		"visible": adm.visible,
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

func (adm *Admin) show(term admin.Terminal, args []string) error {
	for k, v := range adm.mod.Cache().Clone() {
		attachments := v.Status.Attachments.Objects()

		term.Printf("%v:%v - %v (%v), %v objects\n",
			k,
			uint16(v.Status.Port),
			v.Status.Alias,
			v.Identity,
			len(attachments),
		)

		for _, a := range attachments {
			id, err := astral.ResolveObjectID(a)
			if err != nil {
				return err
			}
			term.Printf("- %v (%v)\n", a.ObjectType(), id)
		}
	}

	return nil
}

func (adm *Admin) scan(term admin.Terminal, args []string) error {
	return adm.mod.Scan()
}

func (adm *Admin) update(term admin.Terminal, args []string) error {
	return adm.mod.Broadcast()
}

func (adm *Admin) visible(term admin.Terminal, args []string) (err error) {
	var arg string
	if len(args) > 0 {
		arg = args[0]
	}

	switch arg {
	case "":
		term.Printf("%v\n", adm.mod.visible.Get())
	case "f", "false", "n", "no", "off":
		return adm.mod.SetVisible(false)
	case "t", "true", "y", "yes", "on":
		return adm.mod.SetVisible(true)
	default:
		return errors.New("invalid value")
	}

	return
}

func (adm *Admin) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: %v <command>\n\n", status.ModuleName)
	term.Printf("commands:\n")
	term.Printf("  scan                          broadcast a scan message to collect statuses\n")
	term.Printf("  show                          show cached statuses\n")
	term.Printf("  update                        broadcast a status update\n")
	term.Printf("  visible [bool]                show or set visibility\n")
	term.Printf("  help                          show help\n")
	return nil
}

func (adm *Admin) ShortDescription() string {
	return "list present identities"
}
