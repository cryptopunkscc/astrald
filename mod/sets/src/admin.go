package sets

import (
	"cmp"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"slices"
	"strconv"
	"strings"
)

type Admin struct {
	mod  *Module
	cmds map[string]func(admin.Terminal, []string) error
}

func NewAdmin(mod *Module) *Admin {
	var adm = &Admin{mod: mod}
	adm.cmds = map[string]func(admin.Terminal, []string) error{
		"create":       adm.create,
		"delete":       adm.delete,
		"list":         adm.list,
		"add":          adm.add,
		"remove":       adm.remove,
		"add_union":    adm.addUnion,
		"remove_union": adm.removeUnion,
		"info":         adm.info,
		"show":         adm.show,
		"where":        adm.where,
		"contains":     adm.contains,
		"set_visible":  adm.setVisible,
		"set_desc":     adm.setDesc,
		"help":         adm.help,
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

func (adm *Admin) setDesc(term admin.Terminal, args []string) error {
	if len(args) < 2 {
		return errors.New("missing argument")
	}

	name, desc := args[0], args[1]

	return adm.mod.SetDescription(name, desc)
}

func (adm *Admin) setVisible(term admin.Terminal, args []string) error {
	if len(args) == 0 {
		return errors.New("missing argument")
	}

	var visible = true
	name := args[0]

	if len(args) >= 2 {
		switch args[1] {
		case "t", "true", "y", "yes":
			visible = true
		case "f", "false", "n", "no":
			visible = false
		default:
			return fmt.Errorf("invalid argument: %s", args[1])
		}
	}

	return adm.mod.SetVisible(name, visible)
}

func (adm *Admin) create(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	var typ = sets.TypeSet

	if len(args) >= 2 {
		typ = sets.Type(args[1])
	}

	_, err := adm.mod.CreateSet(args[0], typ)

	return err
}

func (adm *Admin) delete(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	return adm.mod.DeleteSet(args[0])
}

func (adm *Admin) list(term admin.Terminal, _ []string) error {
	list, err := adm.mod.AllSets()
	if err != nil {
		return err
	}

	slices.SortFunc(list, func(a, b sets.Info) int {
		return cmp.Compare(a.Name, b.Name)
	})

	var f = "%-20s %-6s %8s %1s %v\n"
	term.Printf(f, admin.Header("Created at"), admin.Header("Type"), admin.Header("Size"), admin.Header("V"), admin.Header("Name"))
	for _, item := range list {
		var v = "n"
		if item.Visible {
			v = "y"
		}
		name := item.Name
		if item.Description != "" {
			name = item.Description + " (" + item.Name + ")"
		}
		term.Printf(f,
			item.CreatedAt,
			item.Type,
			strconv.Itoa(item.Size),
			v,
			name,
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

func (adm *Admin) addUnion(term admin.Terminal, args []string) error {
	if len(args) < 2 {
		return errors.New("missing argument")
	}

	return adm.mod.AddToUnion(args[0], args[1])
}

func (adm *Admin) removeUnion(term admin.Terminal, args []string) error {
	if len(args) < 2 {
		return errors.New("missing argument")
	}

	return adm.mod.RemoveFromUnion(args[0], args[1])
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

		var status = "yes"
		_, err = adm.mod.Member(name, dataID)
		switch {
		case err == nil:
		case errors.Is(err, sets.ErrMemberNotFound):
			status = "no"
		default:
			term.Printf("%v: %v\n", dataID, err)
			continue
		}

		term.Printf("%v: %v\n", dataID, status)
	}

	return nil
}

func (adm *Admin) show(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	name := args[0]

	list, err := adm.mod.Scan(name, nil)
	if err != nil {
		return err
	}

	var f = "%-20s %-8s %s\n"
	term.Printf("\n")
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

func (adm *Admin) info(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	name := args[0]

	setRow, err := adm.mod.dbFindSetByName(name)
	if err != nil {
		return err
	}

	var size int
	size, err = adm.mod.dbMemberCountBySetID(setRow.ID)

	term.Printf("Set type: %v\n", setRow.Type)
	term.Printf("Created at: %v\n", setRow.CreatedAt)
	term.Printf("Set size: %v\n", size)
	if sets.Type(setRow.Type) == sets.TypeUnion {
		unions, err := adm.mod.dbUnionFindByUnionID(setRow.ID)
		if err != nil {
			return err
		}

		slices.SortFunc(unions, func(a, b dbUnion) int {
			return cmp.Compare(a.Set.Name, b.Set.Name)
		})

		term.Printf("%v\n", admin.Header("Members"))
		for _, union := range unions {
			term.Printf("%v\n", union.Set.Name)
		}
	}

	return nil
}

func (adm *Admin) where(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	dataID, err := data.Parse(args[0])
	if err != nil {
		return err
	}

	list, err := adm.mod.Where(dataID)
	if err != nil {
		return err
	}

	if len(list) == 0 {
		return errors.New("not found")
	}

	slices.Sort(list)

	term.Printf("%s\n", strings.Join(list, "\n"))

	return nil
}

func (adm *Admin) ShortDescription() string {
	return "manage " + sets.ModuleName
}

func (adm *Admin) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: %s <command>\n\n", sets.ModuleName)
	term.Printf("commands:\n")
	term.Printf("  list                          list all sets\n")
	term.Printf("  create <name> [type]          create a new set, type can be set/union (default=set)\n")
	term.Printf("  delete <name>                 delete a set\n")
	term.Printf("  add <name> <dataID>           add data to a set\n")
	term.Printf("  remove <name> <dataID>        remove data from a set\n")
	term.Printf("  add_union <union> <set>       add a set to a union\n")
	term.Printf("  where <dataID>                show sets containing data\n")
	term.Printf("  contains <name> <[]dataID>    check if a set contains data\n")
	term.Printf("  info <name>                   show info about a set\n")
	term.Printf("  show <name>                   show set members\n")
	term.Printf("  help                          show help\n")
	return nil
}
