package index

import (
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/index"
	"github.com/cryptopunkscc/astrald/mod/wallet"
	"strconv"
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
		"create":   adm.create,
		"delete":   adm.delete,
		"list":     adm.list,
		"add":      adm.add,
		"remove":   adm.remove,
		"show":     adm.show,
		"find":     adm.find,
		"contains": adm.contains,
		"help":     adm.help,
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

func (adm *Admin) create(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	_, err := adm.mod.CreateIndex(args[0], index.TypeSet)

	return err
}

func (adm *Admin) delete(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	return adm.mod.DeleteIndex(args[0])
}

func (adm *Admin) list(term admin.Terminal, _ []string) error {
	list, err := adm.mod.AllIndexes()
	if err != nil {
		return err
	}

	var f = "%-20s %-6s %8s %v\n"
	term.Printf(f, admin.Header("Created at"), admin.Header("Type"), admin.Header("Size"), admin.Header("Name"))
	for _, item := range list {
		term.Printf(f,
			item.CreatedAt,
			item.Type,
			strconv.Itoa(item.Size),
			item.Name,
		)
	}

	return nil
}

func (adm *Admin) add(term admin.Terminal, args []string) error {
	if len(args) < 2 {
		return errors.New("missing argument")
	}

	var name = args[0]

	for _, d := range args[1:] {
		dataID, err := data.Parse(d)
		if err != nil {
			term.Printf("%v parse error: %v\n", d, err)
			continue
		}
		err = adm.mod.AddToSet(name, dataID)
		if err != nil {
			term.Printf("%v error: %v\n", dataID, err)
			continue
		}
		term.Printf("%v added\n", dataID)
	}

	return nil
}

func (adm *Admin) contains(term admin.Terminal, args []string) error {
	if len(args) < 2 {
		return errors.New("missing argument")
	}

	var name = args[0]

	for _, d := range args[1:] {
		dataID, err := data.Parse(d)
		if err != nil {
			term.Printf("%v: %v\n", d, fmt.Errorf("parse error: %w", err))
			continue
		}
		c, err := adm.mod.Contains(name, dataID)
		if err != nil {
			term.Printf("%v: %v\n", dataID, err)
			continue
		}

		var s = "yes"
		if !c {
			s = "no"
		}

		term.Printf("%v: %v\n", dataID, s)
	}

	return nil
}

func (adm *Admin) show(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	name := args[0]

	list, err := adm.mod.UpdatedSince(name, time.Time{})
	if err != nil {
		return err
	}

	var f = "%-20s %-8s %s\n"
	term.Printf(f, admin.Header("Updated at"), admin.Header("Status"), admin.Header("DataID"))
	for _, item := range list {
		var status = "added"
		if !item.Added {
			status = "removed"
		}

		term.Printf(f,
			item.UpdatedAt,
			status,
			item.DataID,
		)
	}

	return nil
}

func (adm *Admin) find(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	dataID, err := data.Parse(args[0])
	if err != nil {
		return err
	}

	list, err := adm.mod.Find(dataID)
	if err != nil {
		return err
	}

	if len(list) == 0 {
		return errors.New("not found")
	}

	term.Printf("%s\n", strings.Join(list, "\n"))

	return nil
}

func (adm *Admin) remove(term admin.Terminal, args []string) error {
	if len(args) < 2 {
		return errors.New("missing argument")
	}

	var name = args[0]

	for _, d := range args[1:] {
		dataID, err := data.Parse(d)
		if err != nil {
			term.Printf("%v parse error: %v\n", d, err)
			continue
		}
		err = adm.mod.RemoveFromSet(name, dataID)
		if err != nil {
			term.Printf("%v error: %v\n", dataID, err)
			continue
		}
		term.Printf("%v removed\n", dataID)
	}

	return nil
}

func (adm *Admin) ShortDescription() string {
	return "manage " + wallet.ModuleName
}

func (adm *Admin) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: %s <command>\n\n", wallet.ModuleName)
	term.Printf("commands:\n")
	term.Printf("  help            show help\n")
	return nil
}
