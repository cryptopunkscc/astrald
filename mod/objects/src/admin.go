package objects

import (
	"cmp"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"github.com/cryptopunkscc/astrald/sig"
	"io"
	"reflect"
	"regexp"
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
		"read":     adm.read,
		"purge":    adm.purge,
		"describe": adm.describe,
		"find":     adm.find,
		"fetch":    adm.fetch,
		"info":     adm.info,
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

func (adm *Admin) read(term admin.Terminal, args []string) error {
	var err error
	var opts = &objects.OpenOpts{
		Zone:           objects.DefaultZones,
		IdentityFilter: id.AllowEveryone,
	}
	var zones string

	var flags = flag.NewFlagSet("read", flag.ContinueOnError)
	flags.StringVar(&zones, "z", "lv", "enabled zones")
	flags.SetOutput(term)
	err = flags.Parse(args)
	if err != nil {
		return err
	}

	if len(flags.Args()) == 0 {
		return errors.New("missing data id")
	}

	if len(zones) > 0 {
		opts.Zone = objects.Zones(zones)
	}

	for _, arg := range flags.Args() {
		objectID, err := object.ParseID(arg)
		if err != nil {
			return err
		}

		r, err := adm.mod.Open(context.Background(), objectID, opts)
		if err != nil {
			return err
		}

		io.Copy(term, r)
	}

	return nil
}

func (adm *Admin) describe(term admin.Terminal, args []string) error {
	var err error
	var opts = &desc.Opts{
		IdentityFilter: id.AllowEveryone,
	}

	// parse args
	var flags = flag.NewFlagSet("describe", flag.ContinueOnError)
	flags.BoolVar(&opts.Network, "n", false, "use network sources")
	flags.SetOutput(term)
	err = flags.Parse(args)
	if err != nil {
		return err
	}

	args = flags.Args()

	if len(args) == 0 {
		return errors.New("missing object id")
	}

	objectID, err := object.ParseID(args[0])
	if err != nil {
		return err
	}

	var desc = adm.mod.Describe(context.Background(), objectID, opts)

	term.Printf("%-6s %v\n", admin.Header("SHA256"), admin.Keyword(hex.EncodeToString(objectID.Hash[:])))
	term.Printf("%-6s %v", admin.Header("SIZE"), admin.Keyword(log.DataSize(objectID.Size).HumanReadable()))

	if objectID.Size > 1023 {
		term.Printf(" (%v bytes)", objectID.Size)
	}

	term.Printf("\n\n")

	// print descriptors
	for _, d := range desc {
		term.Printf("%v: %v\n  ", d.Source, admin.Keyword(d.Data.Type()))

		j, err := json.MarshalIndent(d.Data, "  ", "  ")
		if err != nil {
			term.Printf("marshal error: %v\n", err)
		}
		term.Printf("%s\n\n", string(j))
	}

	return nil
}

func (adm *Admin) find(term admin.Terminal, args []string) error {
	if len(args) == 0 {
		return errors.New("missing argument")
	}

	var opts = &objects.FindOpts{}

	matches, err := adm.mod.Find(context.Background(), args[0], opts)

	for _, match := range matches {
		var name string

		if adm.mod.content != nil {
			name = adm.mod.content.BestTitle(match.ObjectID)
		}

		term.Printf("%-64s %v; %v\n",
			match.ObjectID,
			match.Exp,
			name,
		)
	}

	return err
}

func (adm *Admin) fetch(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	term.Printf("fetching %v...\n", args[0])

	objectID, err := adm.mod.Fetch(args[0])

	if err != nil {
		return err
	}

	term.Printf("stored as %v (%s)\n", objectID, log.DataSize(objectID.Size))

	return nil
}

func (adm *Admin) info(term admin.Terminal, args []string) error {
	var f = "%6s %s\n"

	// list openers
	openers := adm.mod.openers.Clone()
	slices.SortFunc(openers, func(a, b *Opener) int {
		return cmp.Compare(a.Priority, b.Priority) * -1
	})

	term.Printf("Openers:\n")
	term.Printf(f, admin.Header("Prio"), admin.Header("Type"))
	for _, opener := range openers {
		term.Printf(
			f,
			strconv.FormatInt(int64(opener.Priority), 10),
			reflect.TypeOf(opener.Opener),
		)
	}
	term.Println()

	// list creators
	creators := adm.mod.creators.Clone()
	slices.SortFunc(creators, func(a, b *Creator) int {
		return cmp.Compare(a.Priority, b.Priority) * -1
	})

	term.Printf("Creators:\n")
	term.Printf(f, admin.Header("Prio"), admin.Header("Type"))
	for _, creator := range creators {
		term.Printf(
			f,
			strconv.FormatInt(int64(creator.Priority), 10),
			reflect.TypeOf(creator.Creator),
		)
	}
	term.Println()

	term.Printf("\n%v\n\n", admin.Header("Describers"))
	list, _ := sig.MapSlice(adm.mod.describers.Clone(), func(i objects.Describer) (string, error) {
		if s, ok := i.(fmt.Stringer); ok {
			return s.String(), nil
		}
		return reflect.TypeOf(i).String(), nil
	})
	slices.Sort(list)

	for _, p := range list {
		term.Printf("%s\n", p)
	}

	term.Printf("\n%v\n\n", admin.Header("Finders"))
	list, _ = sig.MapSlice(adm.mod.finders.Clone(), func(i objects.Finder) (string, error) {
		if s, ok := i.(fmt.Stringer); ok {
			return s.String(), nil
		}
		return reflect.TypeOf(i).String(), nil
	})
	slices.Sort(list)

	for _, p := range list {
		term.Printf("%s\n", p)
	}

	return nil
}

func (adm *Admin) ShortDescription() string {
	return "manage objects"
}

func (adm *Admin) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: objects <command>\n\n")
	term.Printf("commands:\n")
	term.Printf("  read [objectID]                           read an object (caution - may print binary data)\n")
	term.Printf("  fetch <url>                               download an object to storage\n")
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
