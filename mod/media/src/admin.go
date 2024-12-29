package media

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/media"
	"github.com/cryptopunkscc/astrald/object"
)

type Admin struct {
	mod  *Module
	cmds map[string]func(admin.Terminal, []string) error
}

func NewAdmin(mod *Module) *Admin {
	var adm = &Admin{mod: mod}
	adm.cmds = map[string]func(admin.Terminal, []string) error{
		"index":  adm.index,
		"forget": adm.forget,
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

func (adm *Admin) index(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	objectID, err := object.ParseID(args[0])
	if err != nil {
		return err
	}

	audio, err := adm.mod.audio.Index(context.Background(), objectID, nil)
	if audio != nil {
		json.NewEncoder(term).Encode(audio)
	}

	return err
}

func (adm *Admin) forget(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	objectID, err := object.ParseID(args[0])
	if err != nil {
		return err
	}

	return adm.mod.audio.Forget(objectID)
}

func (adm *Admin) ShortDescription() string {
	return "manage " + media.ModuleName
}

func (adm *Admin) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: %v <command>\n\n", media.ModuleName)
	term.Printf("commands:\n")
	term.Printf("  help               show help\n")
	return nil
}
