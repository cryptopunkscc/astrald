package zip

import (
	"errors"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/object"
)

type Admin struct {
	mod  *Module
	cmds map[string]func(admin.Terminal, []string) error
}

func NewAdmin(mod *Module) *Admin {
	var adm = &Admin{mod: mod}
	adm.cmds = map[string]func(admin.Terminal, []string) error{
		"index":   adm.index,
		"unindex": adm.unindex,
		"forget":  adm.forget,
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

func (adm *Admin) index(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	objectID, err := object.ParseID(args[0])
	if err != nil {
		return err
	}

	return adm.mod.Index(objectID)
}

func (adm *Admin) unindex(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	objectID, err := object.ParseID(args[0])
	if err != nil {
		return err
	}

	return adm.mod.Unindex(objectID)
}

func (adm *Admin) forget(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	objectID, err := object.ParseID(args[0])
	if err != nil {
		return err
	}

	return adm.mod.Forget(objectID)
}

func (adm *Admin) ShortDescription() string {
	return "zip indexer"
}

func (adm *Admin) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: zip <command>\n\n")
	term.Printf("commands:\n")
	term.Printf("  index <objectID>           index the contents of a zip file\n")
	term.Printf("  help                       show help\n")
	return nil
}
