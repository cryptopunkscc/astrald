package sets

import (
	"errors"
	"flag"
	"fmt"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"github.com/cryptopunkscc/astrald/node"
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
		"list":        adm.list,
		"create":      adm.create,
		"delete":      adm.delete,
		"add":         adm.add,
		"remove":      adm.remove,
		"include":     adm.include,
		"exclude":     adm.exclude,
		"scan":        adm.scan,
		"sync":        adm.sync,
		"show":        adm.show,
		"where":       adm.where,
		"set_visible": adm.setVisible,
		"set_desc":    adm.setDesc,
		"help":        adm.help,
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

	var typ = sets.TypeBasic

	if len(args) >= 2 {
		typ = sets.Type(args[1])
	}

	_, err := adm.mod.CreateManaged(args[0], typ)

	return err
}

func (adm *Admin) delete(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	set, err := adm.mod.Open(args[0], false)
	if err != nil {
		return err
	}

	return set.Delete()
}

func (adm *Admin) add(term admin.Terminal, args []string) error {
	if len(args) < 2 {
		return errors.New("missing argument")
	}

	set, err := adm.mod.Open(args[0], false)
	if err != nil {
		return err
	}

	for _, arg := range args[1:] {
		if dataID, err := data.Parse(arg); err == nil {
			err = set.Add(dataID)
			if err != nil {
				term.Printf("add %v: %v\n", dataID, err)
			} else {
				term.Printf("add %v: ok\n", dataID)
			}
			continue
		}
		if i, err := strconv.ParseUint(arg, 10, 64); err == nil {
			err = set.AddByID(uint(i))
			if err != nil {
				term.Printf("add %v: %v\n", i, err)
			} else {
				term.Printf("add %v: ok\n", i)
			}
			continue
		}

		term.Printf("add %v: %v\n", arg, errors.New("unrecognized id"))
	}

	return nil
}

func (adm *Admin) remove(term admin.Terminal, args []string) error {
	if len(args) < 2 {
		return errors.New("missing argument")
	}

	set, err := adm.mod.Open(args[0], false)
	if err != nil {
		return err
	}

	for _, arg := range args[1:] {
		if dataID, err := data.Parse(arg); err == nil {
			err = set.Remove(dataID)
			if err != nil {
				term.Printf("remove %v: %v\n", dataID, err)
			} else {
				term.Printf("remove %v: ok\n", dataID)
			}
			continue
		}
		if i, err := strconv.ParseUint(arg, 10, 64); err == nil {
			err = set.RemoveByID(uint(i))
			if err != nil {
				term.Printf("remove %v: %v\n", i, err)
			} else {
				term.Printf("remove %v: ok\n", i)
			}
			continue
		}

		term.Printf("remove %v: %v\n", arg, errors.New("unrecognized id"))
	}

	return nil
}

func (adm *Admin) list(term admin.Terminal, _ []string) error {
	list, err := adm.mod.All()
	if err != nil {
		return err
	}

	slices.Sort(list)

	var f = "%-40s %8s %10s %s\n"
	term.Printf(f, admin.Header("Name"), admin.Header("Count"), admin.Header("Size"), admin.Header("Type"))
	for _, item := range list {
		set, err := adm.mod.Open(item, false)
		if err != nil {
			continue
		}

		stat, err := set.Stat()
		if err != nil {
			continue
		}

		term.Printf(f,
			node.FormatString(adm.mod.node, set.DisplayName()),
			strconv.Itoa(stat.Size),
			log.DataSize(stat.DataSize).HumanReadable(),
			stat.Type,
		)
	}

	return nil
}

func (adm *Admin) sync(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	var name = args[0]

	set, err := adm.mod.Open(name, false)
	if err != nil {
		return err
	}

	switch typed := set.(type) {
	case *Union:
		return typed.Sync()
	}

	return errors.New("unsupported set type")
}

func (adm *Admin) include(term admin.Terminal, args []string) error {
	if len(args) < 2 {
		return errors.New("missing argument")
	}

	super, err := adm.mod.OpenUnion(args[0], false)
	if err != nil {
		return err
	}

	return super.AddSubset(args[1:]...)
}

func (adm *Admin) exclude(term admin.Terminal, args []string) error {
	if len(args) < 2 {
		return errors.New("missing argument")
	}

	super, err := adm.mod.OpenUnion(args[0], false)
	if err != nil {
		return err
	}

	return super.RemoveSubset(args[1:]...)
}

func (adm *Admin) scan(term admin.Terminal, args []string) error {
	opts := &sets.ScanOpts{}

	var flags = flag.NewFlagSet("scan", flag.ContinueOnError)
	flags.BoolVar(&opts.IncludeRemoved, "r", false, "show removed objects")
	err := flags.Parse(args)
	if err != nil {
		return err
	}

	if len(flags.Args()) < 1 {
		return errors.New("set name missing")
	}
	name := flags.Args()[0]

	list, err := adm.mod.Scan(name, opts)
	if err != nil {
		return err
	}

	var f = "%-20s %-8s %s\n"
	term.Printf("\n")
	term.Printf(f, admin.Header("Updated at"), admin.Header("Removed"), admin.Header("DataID"))
	for _, item := range list {
		var status = "no"
		if item.Removed {
			status = "yes"
		}

		term.Printf(f,
			item.UpdatedAt,
			status,
			item.DataID,
		)
	}

	return nil
}

func (adm *Admin) show(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	name := args[0]

	set, err := adm.mod.Open(name, false)
	if err != nil {
		return err
	}

	info, err := set.Stat()
	if err != nil {
		return err
	}

	term.Printf("SetScanner type: %v\n", info.Type)
	term.Printf("Created at: %v\n", info.CreatedAt)
	term.Printf("SetScanner size: %v\n", info.Size)

	switch typed := set.(type) {
	case *Union:
		term.Printf("Subsets:\n")
		subsets, err := typed.Subsets()
		if err != nil {
			return err
		}
		for _, s := range subsets {
			term.Printf("- %s\n", s)
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

func (adm *Admin) ShortDescription() string {
	return "manage " + sets.ModuleName
}

func (adm *Admin) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: %s <command>\n\n", sets.ModuleName)
	term.Printf("commands:\n")
	term.Printf("  list                          list all sets\n")
	term.Printf("  create <name> [type]          create a new set (default type=basic)\n")
	term.Printf("  delete <name>                 delete a set\n")
	term.Printf("  add <name> <dataID>           add data to a set\n")
	term.Printf("  remove <name> <dataID>        remove data from a set\n")
	term.Printf("  include <superset> <subset>   add a set to a union\n")
	term.Printf("  exclude <superset> <subset>   remove a set from a union\n")
	term.Printf("  scan [-r] <set>               list objects in a set; use -r to include removed\n")
	term.Printf("  where <dataID>                show sets containing data\n")
	term.Printf("  show <name>                   show info about a set\n")
	term.Printf("  help                          show help\n")
	return nil
}
