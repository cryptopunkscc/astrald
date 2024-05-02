package sets

import (
	"errors"
	"flag"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"github.com/cryptopunkscc/astrald/object"
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
		"list":   adm.list,
		"create": adm.create,
		"delete": adm.delete,
		"add":    adm.add,
		"remove": adm.remove,
		"scan":   adm.scan,
		"show":   adm.show,
		"where":  adm.where,
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

func (adm *Admin) create(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	_, err := adm.mod.Create(args[0])

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
		objectID, err := object.ParseID(arg)
		if err != nil {
			term.Printf("%v: parse error: %v\n", arg, err)
			continue
		}

		err = set.Add(objectID)
		if err != nil {
			term.Printf("%v: add error: %v\n", objectID, err)
		} else {
			term.Printf("%v: added\n", objectID)
		}
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
		objectID, err := object.ParseID(arg)
		if err == nil {
			term.Printf("%v: parse error: %v\n", arg, err)
			continue
		}

		err = set.Remove(objectID)
		if err != nil {
			term.Printf("%v: remove error: %v\n", objectID, err)
		} else {
			term.Printf("%v: removed\n", objectID)
		}
	}

	return nil
}

func (adm *Admin) list(term admin.Terminal, _ []string) error {
	list, err := adm.mod.All()
	if err != nil {
		return err
	}

	slices.Sort(list)

	var f = "%-40s %8s %10s\n"
	term.Printf(f, admin.Header("Name"), admin.Header("Count"), admin.Header("Size"))
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
			set.Name(),
			strconv.Itoa(stat.Size),
			log.DataSize(stat.DataSize).HumanReadable(),
		)
	}

	return nil
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
	term.Printf(f, admin.Header("Updated at"), admin.Header("Removed"), admin.Header("ObjectID"))
	for _, item := range list {
		var status = "no"
		if item.Removed {
			status = "yes"
		}

		term.Printf(f,
			item.UpdatedAt,
			status,
			item.ObjectID,
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

	term.Printf("Created at: %v\n", info.CreatedAt)
	term.Printf("Set size: %v\n", info.Size)

	return nil
}

func (adm *Admin) where(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	objectID, err := object.ParseID(args[0])
	if err != nil {
		return err
	}

	list, err := adm.mod.Where(objectID)
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
	term.Printf("  create <name> [type]          create a new set (default type=basic)\n")
	term.Printf("  delete <name>                 delete a set\n")
	term.Printf("  add <name> <objectID>         add an object to a set\n")
	term.Printf("  remove <name> <objectID>      remove an object from a set\n")
	term.Printf("  scan [-r] <set>               list objects in a set; use -r to include removed\n")
	term.Printf("  show <name>                   show info about a set\n")
	term.Printf("  where <objectID>              show sets containing the object\n")
	term.Printf("  help                          show help\n")
	return nil
}
