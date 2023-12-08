package fs

import (
	"errors"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin/api"
)

type Admin struct {
	mod  *Module
	cmds map[string]func(admin.Terminal, []string) error
}

func NewAdmin(mod *Module) *Admin {
	var adm = &Admin{mod: mod}
	adm.cmds = map[string]func(admin.Terminal, []string) error{
		"add":  adm.add,
		"ls":   adm.ls,
		"find": adm.find,
		"info": adm.info,
		"help": adm.help,
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

func (adm *Admin) add(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	return adm.mod.index.Add(args[0])
}

func (adm *Admin) ls(term admin.Terminal, args []string) error {
	var localFiles []*dbLocalFile
	tx := adm.mod.db.Order("indexed_at").Find(&localFiles)
	if tx.Error != nil {
		return tx.Error
	}

	var f = "%-64s %-20s %s\n"
	term.Printf(f, admin.Header("DataID"), admin.Header("Indexed"), admin.Header("Path"))
	for _, localFile := range localFiles {
		dataID, err := data.Parse(localFile.DataID)
		if err != nil {
			continue
		}

		term.Printf(f, dataID, localFile.IndexedAt, localFile.Path)
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

	files := adm.mod.dbFindByID(dataID)
	term.Printf("found %d path(s)\n", len(files))
	for _, file := range files {
		term.Printf("%s\n", file.Path)
	}

	return nil
}

func (adm *Admin) info(term admin.Terminal, args []string) error {
	f := "%-64s %s\n"

	term.Printf(f, admin.Header("Store Path"), admin.Header("Free"))
	for _, path := range adm.mod.store.Paths() {
		var free int
		usage, _ := DiskUsage(path)
		if usage != nil {
			free = int(usage.Free)
		}

		term.Printf(f, path, log.DataSize(free))
	}

	term.Printf("\n%s\n", admin.Header("INDEX PATH"))
	for _, path := range adm.mod.index.watcher.WatchList() {
		term.Printf("%s\n", path)
	}
	return nil
}

func (adm *Admin) ShortDescription() string {
	return "manage local filesystem"
}

func (adm *Admin) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: fs <command>\n\n")
	term.Printf("commands:\n")
	term.Printf("  add <path>                 add a path to the index\n")
	term.Printf("  ls                         list all indexed files\n")
	term.Printf("  find <dataID>              show local path(s) of the data\n")
	term.Printf("  info                       show index info\n")
	term.Printf("  help                       show help\n")
	return nil
}
