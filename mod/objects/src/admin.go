package objects

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/astral/term"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"regexp"
	"slices"
	"strings"
)

type Admin struct {
	mod  *Module
	cmds map[string]func(admin.Terminal, []string) error
}

func NewAdmin(mod *Module) *Admin {
	var adm = &Admin{mod: mod}
	adm.cmds = map[string]func(admin.Terminal, []string) error{
		"blueprints": adm.blueprints,
		"fetch":      adm.fetch,
		"holders":    adm.holders,
		"info":       adm.info,
		"purge":      adm.purge,
		"push":       adm.push,
		"help":       adm.help,
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

func (adm *Admin) blueprints(term admin.Terminal, args []string) error {
	types := adm.mod.blueprints.Names()

	slices.Sort(types)

	term.Printf("%v blueprints:\n", len(types))

	for _, t := range types {
		term.Printf("- %v\n", t)
	}

	return nil
}

func (adm *Admin) holders(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return adm.help(term, []string{})
	}

	objectID, err := object.ParseID(args[0])
	if err != nil {
		return err
	}

	for _, h := range adm.mod.Holders(objectID) {
		term.Printf("%v\n", h)
	}

	return nil
}

func (adm *Admin) purge(term admin.Terminal, args []string) error {
	if len(args) == 0 {
		return errors.New("missing argument")
	}

	for _, arg := range args {
		objectID, err := object.ParseID(arg)
		if err != nil {
			term.Printf("parse '%v': %v\n", arg, err)
		}

		n, err := adm.mod.Purge(objectID, nil)

		var extra string
		if err != nil {
			extra = " (with errors)"
		}

		term.Printf("%v: purged %v%v\n", objectID, n, extra)
	}

	return nil
}

func (adm *Admin) fetch(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	term.Printf("fetching %v...\n", args[0])

	objectID, err := adm.mod.fetch(args[0])

	if err != nil {
		return err
	}

	term.Printf("stored as %v (%v)\n", objectID, log.DataSize(objectID.Size))

	return nil
}

func (adm *Admin) info(term admin.Terminal, args []string) error {
	term.Printf("Describers: %v\n", strings.Join(strSort(adm.mod.describers.Clone()), ", "))
	term.Printf("Purger:     %v\n", strings.Join(strSort(adm.mod.purgers.Clone()), ", "))
	term.Printf("Searcher:   %v\n", strings.Join(strSort(adm.mod.searchers.Clone()), ", "))
	term.Printf("Finder:     %v\n", strings.Join(strSort(adm.mod.finders.Clone()), ", "))
	term.Printf("Holder:     %v\n", strings.Join(strSort(adm.mod.holders.Clone()), ", "))
	term.Printf("Receiver:   %v\n", strings.Join(strSort(adm.mod.receivers.Clone()), ", "))

	return nil
}

func strSort[T any](a []T) (s []string) {
	s = term.StringifySlice(a)
	slices.Sort(s)
	return
}

func (adm *Admin) push(term admin.Terminal, args []string) error {
	if len(args) < 2 {
		return errors.New("missing arguments")
	}

	target, err := adm.mod.Dir.ResolveIdentity(args[0])
	if err != nil {
		return err
	}

	objectID, err := object.ParseID(args[1])
	if err != nil {
		return err
	}

	lctx := astral.NewContext(nil).WithIdentity(term.UserIdentity())

	obj, err := objects.Load[astral.Object](lctx, adm.mod.Root(), objectID, adm.mod.Blueprints())
	if err != nil {
		return err
	}

	ctx := astral.NewContext(nil).WithIdentity(adm.mod.node.Identity())

	return adm.mod.Push(ctx, target, obj)
}

func (adm *Admin) ShortDescription() string {
	return "manage objects"
}

func (adm *Admin) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: objects <command>\n\n")
	term.Printf("commands:\n")
	term.Printf("  read [objectID]                           read an object (caution - may print binary data)\n")
	term.Printf("  fetch <url>                               download an object to storage\n")
	term.Printf("  blueprints                                list all registered blueprints\n")
	term.Printf("  info                                      show info\n")
	term.Printf("  help                                      show help\n")
	return nil
}

func isURL(url string) bool {
	matched, _ := regexp.Match("^https?://", []byte(url))
	return matched
}

func isARL(s string) bool {
	return strings.HasPrefix(s, "astral://")
}
