package objects

import (
	"cmp"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/astral"
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
		"search":   adm.search,
		"fetch":    adm.fetch,
		"hold":     adm.hold,
		"release":  adm.release,
		"holders":  adm.holders,
		"inv":      adm.inv,
		"show":     adm.show,
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

		if holders := adm.mod.Holders(objectID); len(holders) > 0 {
			term.Printf("%v: held by", objectID)
			for _, h := range holders {
				term.Printf(" %v", h)
			}
			term.Printf("\n")
			continue
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
	var opts = objects.DefaultOpenOpts()
	var zones string

	var flags = flag.NewFlagSet("read", flag.ContinueOnError)
	flags.StringVar(&zones, "z", opts.Zone.String(), "enabled zones")
	flags.SetOutput(term)
	err = flags.Parse(args)
	if err != nil {
		return err
	}

	if len(flags.Args()) == 0 {
		return errors.New("missing data id")
	}

	if len(zones) > 0 {
		opts.Zone = astral.Zones(zones)
	}

	for _, arg := range flags.Args() {
		objectID, err := object.ParseID(arg)
		if err != nil {
			return err
		}

		r, err := adm.mod.OpenAs(context.Background(), term.UserIdentity(), objectID, opts)
		if err != nil {
			return err
		}

		io.Copy(term, r)
	}

	return nil
}

func (adm *Admin) show(term admin.Terminal, args []string) error {
	var err error
	var scope = astral.DefaultScope()
	var zones string

	var flags = flag.NewFlagSet("show", flag.ContinueOnError)
	flags.StringVar(&zones, "z", scope.Zone.String(), "enabled zones")
	flags.SetOutput(term)
	err = flags.Parse(args)
	if err != nil {
		return err
	}

	if len(flags.Args()) == 0 {
		return errors.New("missing data id")
	}

	if len(zones) > 0 {
		scope.Zone = astral.Zones(zones)
	}

	for _, arg := range flags.Args() {
		objectID, err := object.ParseID(arg)
		if err != nil {
			return err
		}

		obj, err := adm.mod.Load(context.Background(), objectID, scope)
		if err != nil {
			return err
		}

		term.Printf("%v %s\n\n", objectID, obj.ObjectType())
		j, err := json.MarshalIndent(obj, "  ", "  ")
		if err != nil {
			term.Printf("error encoding to JSON: %v\n", err)
			continue
		}

		term.Printf("%s\n", string(j))
	}

	return nil
}

func (adm *Admin) describe(term admin.Terminal, args []string) error {
	var err error
	var zonesArg string
	var provider string
	var opts = desc.DefaultOpts()

	// parse args
	var flags = flag.NewFlagSet("describe", flag.ContinueOnError)
	flags.StringVar(&zonesArg, "z", opts.Zone.String(), "set zones to use")
	flags.StringVar(&provider, "p", "", "query this provider")
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

	if zonesArg != "" {
		opts.Zone = astral.Zones(zonesArg)
	}

	var descs []*desc.Desc

	if len(provider) > 0 {
		providerID, err := adm.mod.dir.Resolve(provider)
		if err != nil {
			return err
		}

		c := NewConsumer(adm.mod, term.UserIdentity(), providerID)

		descs, err = c.Describe(context.Background(), objectID, desc.DefaultOpts())
		if err != nil {
			return err
		}
	} else {
		descs = adm.mod.Describe(context.Background(), objectID, opts)
	}

	term.Printf("%-6s %v\n", admin.Header("SHA256"), admin.Keyword(hex.EncodeToString(objectID.Hash[:])))
	term.Printf("%-6s %v", admin.Header("SIZE"), admin.Keyword(log.DataSize(objectID.Size).HumanReadable()))

	if objectID.Size > 1023 {
		term.Printf(" (%v bytes)", objectID.Size)
	}

	term.Printf("\n\n")

	// print descriptors
	for _, d := range descs {
		term.Printf("%v: %v\n  ", d.Source, admin.Keyword(d.Data.Type()))

		j, err := json.MarshalIndent(d.Data, "  ", "  ")
		if err != nil {
			term.Printf("marshal error: %v\n", err)
		}
		term.Printf("%s\n\n", string(j))
	}

	return nil
}

func (adm *Admin) search(term admin.Terminal, args []string) error {
	if len(args) == 0 {
		return errors.New("missing argument")
	}

	var opts = objects.DefaultSearchOpts()
	var zonesArg string
	var provider string
	var err error

	var flags = flag.NewFlagSet("describe", flag.ContinueOnError)
	flags.StringVar(&zonesArg, "z", opts.Zone.String(), "set zones to use")
	flags.StringVar(&provider, "p", "", "query this provider")
	flags.SetOutput(term)
	err = flags.Parse(args)
	if err != nil {
		return err
	}

	if zonesArg != "" {
		opts.Zone = astral.Zones(zonesArg)
	}

	args = flags.Args()

	var matches []objects.Match

	if len(provider) > 0 {
		var providerID id.Identity

		providerID, err = adm.mod.dir.Resolve(provider)
		if err != nil {
			return err
		}

		c := NewConsumer(adm.mod, term.UserIdentity(), providerID)

		matches, err = c.Search(context.Background(), args[0])
	} else {
		matches, err = adm.mod.Search(context.Background(), args[0], opts)
	}

	if err != nil {
		return err
	}

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

	objectID, err := adm.mod.fetch(args[0])

	if err != nil {
		return err
	}

	term.Printf("stored as %v (%s)\n", objectID, log.DataSize(objectID.Size))

	adm.mod.Hold(term.UserIdentity(), objectID)

	return nil
}

func (adm *Admin) holders(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	objectID, err := object.ParseID(args[0])
	if err != nil {
		return err
	}

	holderIDs := adm.mod.Holders(objectID)
	for _, holderID := range holderIDs {
		term.Printf("%v\n", holderID)
	}

	return nil
}

func (adm *Admin) inv(term admin.Terminal, args []string) (err error) {
	holderID := term.UserIdentity()

	if len(args) > 0 {
		holderID, err = adm.mod.dir.Resolve(args[0])
		if err != nil {
			return
		}
	}

	objectIDs := adm.mod.Holdings(holderID)
	for _, objectID := range objectIDs {
		term.Printf("%v\n", objectID)
	}

	return nil
}

func (adm *Admin) hold(term admin.Terminal, args []string) error {
	if len(args) < 2 {
		return errors.New("argument missing")
	}

	holderID, err := adm.mod.dir.Resolve(args[0])
	if err != nil {
		return err
	}

	var objectsIDs []object.ID

	for _, arg := range args[1:] {
		objectID, err := object.ParseID(arg)
		if err != nil {
			return fmt.Errorf("invalid object id: %v", arg)
		}
		objectsIDs = append(objectsIDs, objectID)
	}

	return adm.mod.Hold(holderID, objectsIDs...)
}

func (adm *Admin) release(term admin.Terminal, args []string) error {
	if len(args) < 2 {
		return errors.New("argument missing")
	}

	holderID, err := adm.mod.dir.Resolve(args[0])
	if err != nil {
		return err
	}

	var objectsIDs []object.ID

	for _, arg := range args[1:] {
		objectID, err := object.ParseID(arg)
		if err != nil {
			return fmt.Errorf("invalid object id: %v", arg)
		}
		objectsIDs = append(objectsIDs, objectID)
	}

	return adm.mod.Release(holderID, objectsIDs...)
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

	term.Printf("\n%v\n\n", admin.Header("Searchers"))
	list, _ = sig.MapSlice(adm.mod.searchers.Clone(), func(i objects.Searcher) (string, error) {
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
