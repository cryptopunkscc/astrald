package index

import (
	"cmp"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/index"
	"slices"
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
		"create":       adm.create,
		"delete":       adm.delete,
		"list":         adm.list,
		"add":          adm.add,
		"remove":       adm.remove,
		"add_union":    adm.addUnion,
		"remove_union": adm.removeUnion,
		"info":         adm.info,
		"show":         adm.show,
		"find":         adm.find,
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

	var typ = index.TypeSet

	if len(args) >= 2 {
		typ = index.Type(args[1])
	}

	_, err := adm.mod.CreateIndex(args[0], typ)

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

	slices.SortFunc(list, func(a, b index.Info) int {
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

	list, err := adm.mod.UpdatedBetween(name, time.Time{}, time.Time{})
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

	indexRow, err := adm.mod.dbFindIndexByName(name)
	if err != nil {
		return err
	}

	var size int
	size, err = adm.mod.dbEntryCountByIndexID(indexRow.ID)

	term.Printf("Index type: %v\n", indexRow.Type)
	term.Printf("Created at: %v\n", indexRow.CreatedAt)
	term.Printf("Index size: %v\n", size)
	if index.Type(indexRow.Type) == index.TypeUnion {
		unions, err := adm.mod.dbUnionFindByUnionID(indexRow.ID)
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

	slices.Sort(list)

	term.Printf("%s\n", strings.Join(list, "\n"))

	return nil
}

func (adm *Admin) ShortDescription() string {
	return "manage " + index.ModuleName
}

func (adm *Admin) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: %s <command>\n\n", index.ModuleName)
	term.Printf("commands:\n")
	term.Printf("  list                          list all indexes\n")
	term.Printf("  create <name> [type]          create a new index, type can be set/union (default=set)\n")
	term.Printf("  delete <name>                 delete an index\n")
	term.Printf("  add <name> <dataID>           add data to a set\n")
	term.Printf("  remove <name> <dataID>        remove data from a set\n")
	term.Printf("  add_union <union> <set>       add a set to an union\n")
	term.Printf("  find <dataID>                 search indexes for data\n")
	term.Printf("  contains <name> <[]dataID>    check if an index contains provided data\n")
	term.Printf("  info <name>                   info info about an index\n")
	term.Printf("  show <name>                   list all data from an index\n")
	term.Printf("  help                          info help\n")
	return nil
}
