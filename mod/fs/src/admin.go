package fs

import (
	"errors"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/object"
	"slices"
)

type Admin struct {
	mod  *Module
	cmds map[string]func(admin.Terminal, []string) error
}

func NewAdmin(mod *Module) *Admin {
	var adm = &Admin{mod: mod}
	adm.cmds = map[string]func(admin.Terminal, []string) error{
		"watch":  adm.watch,
		"update": adm.update,
		"rename": adm.rename,
		"find":   adm.find,
		"path":   adm.path,
		"info":   adm.info,
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

func (adm *Admin) watch(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	added, err := adm.mod.Watch(args[0])
	if err != nil {
		return err
	}

	for _, path := range added {
		term.Printf("%v\n", path)
	}

	return nil
}

func (adm *Admin) update(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	objectID, err := adm.mod.update(args[0])
	if err != nil {
		return err
	}

	term.Printf("id: %v\n", objectID)

	return nil
}

func (adm *Admin) rename(term admin.Terminal, args []string) error {
	if len(args) < 2 {
		return errors.New("missing argument")
	}

	return adm.mod.rename(args[0], args[1])
}

func (adm *Admin) find(term admin.Terminal, args []string) error {
	files := adm.mod.Find(nil)

	var f = "%v %v\n"
	term.Printf(f, admin.Header("ID"), admin.Header("Path"))
	for _, file := range files {
		term.Printf(f, file.ObjectID, file.Path)
	}

	return nil
}

func (adm *Admin) path(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	objectID, err := object.ParseID(args[0])
	if err != nil {
		return err
	}

	paths := adm.mod.path(objectID)
	term.Printf("found %v path(s)\n", len(paths))
	for _, path := range paths {
		term.Printf("%v\n", path)
	}

	return nil
}

func (adm *Admin) info(term admin.Terminal, args []string) error {
	term.Printf("\n%v\n", admin.Header("Watch Path"))
	paths := adm.mod.watcher.List()

	slices.Sort(paths)

	for _, path := range paths {
		term.Printf("%v\n", path)
	}
	return nil
}

func (adm *Admin) ShortDescription() string {
	return "manage local filesystem"
}

func (adm *Admin) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: fs <command>\n\n")
	term.Printf("commands:\n")
	term.Printf("  watch <path>               watch a directory tree for changes\n")
	term.Printf("  find                       list all indexed files\n")
	term.Printf("  path <objectID>            show local path(s) for the object\n")
	term.Printf("  info                       show index info\n")
	term.Printf("  help                       show help\n")
	return nil
}
