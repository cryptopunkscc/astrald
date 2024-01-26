package media

import (
	"encoding/json"
	"errors"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/media"
)

type Admin struct {
	mod  *Module
	cmds map[string]func(admin.Terminal, []string) error
}

func NewAdmin(mod *Module) *Admin {
	var adm = &Admin{mod: mod}
	adm.cmds = map[string]func(admin.Terminal, []string) error{
		"scan":  adm.scan,
		"index": adm.index,
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

func (adm *Admin) scan(term admin.Terminal, args []string) error {
	if len(args) == 0 {
		return errors.New("missing argument")
	}

	dataID, err := data.Parse(args[0])
	if err != nil {
		return err
	}

	info, err := adm.mod.indexer.scan(dataID)
	if err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}

	term.Write(bytes)
	term.Println()

	return nil
}

func (adm *Admin) index(term admin.Terminal, args []string) error {
	if len(args) == 0 {
		return errors.New("missing argument")
	}

	dataID, err := data.Parse(args[0])
	if err != nil {
		return err
	}

	info, err := adm.mod.indexer.indexData(dataID)
	if err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}

	term.Write(bytes)
	term.Println()

	return nil
}

func (adm *Admin) ShortDescription() string {
	return "manage " + media.ModuleName
}

func (adm *Admin) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: %s <command>\n\n", media.ModuleName)
	term.Printf("commands:\n")
	term.Printf("  scan [dataID]      scan data and show results\n")
	term.Printf("  index [dataID]     scan data and add to the index\n")
	term.Printf("  help               show help\n")
	return nil
}
