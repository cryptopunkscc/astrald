package objects

import (
	"cmp"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
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
		"blueprints": adm.blueprints,
		"describe":   adm.describe,
		"fetch":      adm.fetch,
		"holders":    adm.holders,
		"info":       adm.info,
		"purge":      adm.purge,
		"push":       adm.push,
		"read":       adm.read,
		"scan":       adm.scan,
		"search":     adm.search,
		"show":       adm.show,
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

	term.Printf("%d blueprints:\n", len(types))

	for _, t := range types {
		term.Printf("- %s\n", t)
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
		term.Printf("%s\n", h)
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

func (adm *Admin) scan(term admin.Terminal, args []string) error {
	var repoName = "default"

	if len(args) > 0 {
		repoName = args[0]
	}

	repo, ok := adm.mod.repos.Get(repoName)
	if !ok {
		return fmt.Errorf("no such repository: %s", repoName)
	}

	scan, err := repo.Scan()
	if err != nil {
		return err
	}

	for id := range scan {
		term.Printf("%s\n", id)
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

		r, err := adm.mod.Open(context.Background(), objectID, &objects.OpenOpts{
			Zone: astral.Zones(zones),
		})
		if err != nil {
			return err
		}

		obj, _, err := adm.mod.Blueprints().Read(r, true)
		if err != nil {
			return err
		}

		term.Printf("%v %s\n\n", objectID, obj.ObjectType())
		j, err := json.MarshalIndent(obj, "  ", "  ")
		if err != nil {
			term.Printf("error encoding to JSON: %v\n", err)
			continue
		}

		term.Printf("  %s\n", string(j))
	}

	return nil
}

func (adm *Admin) describe(term admin.Terminal, args []string) error {
	var err error
	var zonesArg string
	var provider string
	var scope = astral.DefaultScope()

	// parse args
	var flags = flag.NewFlagSet("describe", flag.ContinueOnError)
	flags.StringVar(&zonesArg, "z", scope.Zone.String(), "set zones to use")
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
		scope.Zone = astral.Zones(zonesArg)
	}

	var descs <-chan *objects.SourcedObject

	if len(provider) > 0 {
		var providerID *astral.Identity
		providerID, err = adm.mod.Dir.ResolveIdentity(provider)
		if err != nil {
			return err
		}

		c := NewConsumer(adm.mod, term.UserIdentity(), providerID)

		descs, err = c.Describe(context.Background(), objectID, scope)
	} else {
		descs, err = adm.mod.Describe(context.Background(), objectID, scope)
	}
	if err != nil {
		return err
	}

	term.Printf("%-6s %v\n", admin.Header("SHA256"), admin.Keyword(hex.EncodeToString(objectID.Hash[:])))
	term.Printf("%-6s %v", admin.Header("SIZE"), admin.Keyword(log.DataSize(objectID.Size).HumanReadable()))

	if objectID.Size > 1023 {
		term.Printf(" (%v bytes)", objectID.Size)
	}

	term.Printf("\n\n")

	// print descriptors
	for d := range descs {
		term.Printf("%v: %v\n  ", d.Source, admin.Keyword(d.Object.ObjectType()))

		j, err := json.MarshalIndent(d.Object, "  ", "  ")
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

	var matches <-chan *objects.SearchResult

	if len(provider) > 0 {
		var providerID *astral.Identity

		providerID, err = adm.mod.Dir.ResolveIdentity(provider)
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

	for match := range matches {
		term.Printf("%-64s\n",
			match.ObjectID,
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
	term.Printf(f, admin.Header("Prio"), admin.Header("ObjectType"))
	for _, opener := range openers {
		term.Printf(
			f,
			strconv.FormatInt(int64(opener.Priority), 10),
			reflect.TypeOf(opener.Opener),
		)
	}
	term.Println()

	term.Printf("Describers: %s\n", strings.Join(strSort(adm.mod.describers.Clone()), ", "))
	term.Printf("Purger:     %s\n", strings.Join(strSort(adm.mod.purgers.Clone()), ", "))
	term.Printf("Searcher:   %s\n", strings.Join(strSort(adm.mod.searchers.Clone()), ", "))
	term.Printf("Finder:     %s\n", strings.Join(strSort(adm.mod.finders.Clone()), ", "))
	term.Printf("Holder:     %s\n", strings.Join(strSort(adm.mod.holders.Clone()), ", "))
	term.Printf("Receiver:   %s\n", strings.Join(strSort(adm.mod.receivers.Clone()), ", "))

	term.Printf("\nRepositories:\n")
	for _, repo := range adm.mod.repos.Clone() {
		term.Printf("%v (%v free)\n", repo.Name(), log.DataSize(repo.Free()))
	}

	return nil
}

func strSort[T any](a []T) (s []string) {
	s = log.StringifySlice(a)
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

	obj, err := objects.Load[astral.Object](context.Background(), adm.mod, objectID, astral.DefaultScope())
	if err != nil {
		return err
	}

	return adm.mod.Push(context.Background(), nil, target, obj)
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
