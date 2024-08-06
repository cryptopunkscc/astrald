package keys

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/object"
)

type Admin struct {
	mod  *Module
	cmds map[string]func(admin.Terminal, []string) error
}

func NewAdmin(mod *Module) *Admin {
	var adm = &Admin{mod: mod}
	adm.cmds = map[string]func(admin.Terminal, []string) error{
		"index": adm.index,
		"list":  adm.list,
		"new":   adm.new,
		"help":  adm.help,
	}

	return adm
}

func (adm *Admin) new(term admin.Terminal, args []string) error {
	key, err := astral.GenerateIdentity()
	if err != nil {
		return err
	}

	objectID, err := adm.mod.SaveKey(key)
	if err != nil {
		return err
	}

	if len(args) >= 1 {
		alias := args[0]
		if err := adm.mod.Dir.SetAlias(key, alias); err != nil {
			term.Printf("cannot set alias: %v\n", err)
		}
	}

	term.Printf("created key %s (%s) objectID %v\n",
		key,
		admin.Faded(key.String()),
		objectID,
	)

	return nil
}

func (adm *Admin) list(term admin.Terminal, args []string) error {
	var rows []dbPrivateKey
	tx := adm.mod.db.Find(&rows)
	if tx.Error != nil {
		return tx.Error
	}

	term.Printf("Found %d key(s)\n", len(rows))
	for _, row := range rows {
		term.Printf("%-24s %-64s %v\n", admin.Keyword(row.Type), row.DataID, row.PublicKey)
	}

	return nil
}

func (adm *Admin) index(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	objectID, err := object.ParseID(args[0])
	if err != nil {
		return err
	}

	return adm.mod.IndexKey(objectID)
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

func (adm *Admin) ShortDescription() string {
	return "manage keys"
}

func (adm *Admin) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: %s <command>\n\n", keys.ModuleName)
	term.Printf("commands:\n")
	term.Printf("  new <alias>     create new key with provided alias\n")
	term.Printf("  list            list all keys\n")
	term.Printf("  help            show help\n")
	return nil
}
