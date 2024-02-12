package storage

import (
	"cmp"
	"errors"
	"flag"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/net"
	"io"
	"net/http"
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
		"read":  adm.read,
		"purge": adm.purge,
		"fetch": adm.fetch,
		"info":  adm.info,
		"help":  adm.help,
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
		dataID, err := data.Parse(arg)
		if err != nil {
			term.Printf("parse '%v': %v\n", arg, err)
		}

		n, err := adm.mod.Purge(dataID, nil)

		var extra string
		if err != nil {
			extra = " (with errors)"
		}

		term.Printf("%v: purged %v%v\n", dataID, n, extra)
	}

	return nil
}

func (adm *Admin) read(term admin.Terminal, args []string) error {
	var err error
	var opts = &storage.OpenOpts{
		Virtual:        true,
		Network:        false,
		IdentityFilter: id.AllowEveryone,
	}

	var flags = flag.NewFlagSet("read", flag.ContinueOnError)
	flags.BoolVar(&opts.Virtual, "v", true, "use virtual sources")
	flags.BoolVar(&opts.Network, "n", false, "use network sources")
	flags.SetOutput(term)
	err = flags.Parse(args)
	if err != nil {
		return err
	}

	if len(flags.Args()) == 0 {
		return errors.New("missing data id")
	}

	for _, idstr := range flags.Args() {
		dataID, err := data.Parse(idstr)
		if err != nil {
			return err
		}

		r, err := adm.mod.Open(dataID, opts)
		if err != nil {
			return err
		}

		io.Copy(term, r)
	}

	return nil
}

func (adm *Admin) fetch(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	var url = args[0]

	if matched, _ := regexp.Match("^https?:", []byte(url)); matched {
		term.Printf("downloading %v...\n", url)

		// Make a GET request to the URL
		response, err := http.Get(url)
		if err != nil {
			return err
		}
		defer response.Body.Close()

		var alloc = max(response.ContentLength, 0)

		w, err := adm.mod.Create(
			&storage.CreateOpts{
				Alloc: int(alloc),
			},
		)
		if err != nil {
			return err
		}
		defer w.Discard()

		_, err = io.Copy(w, response.Body)
		if err != nil {
			return err
		}

		dataID, err := w.Commit()
		if err != nil {
			return err
		}

		term.Printf("stored as %v (%s)\n", dataID, log.DataSize(dataID.Size))

		return nil
	}

	if cut, found := strings.CutPrefix(url, "astral:"); found {
		url = cut
	}
	uri, err := adm.mod.Parse(url)
	if err != nil {
		return err
	}

	if uri.User.IsZero() {
		uri.User = term.UserIdentity()
	}

	term.Printf("fetching %v@%v:%v...\n", uri.User, uri.Target, uri.Query)

	var query = net.NewQuery(uri.User, uri.Target, uri.Query)

	conn, err := net.Route(adm.mod.ctx, adm.mod.node.Router(), query)
	if err != nil {
		return err
	}

	w, err := adm.mod.Create(nil)
	if err != nil {
		return err
	}
	defer w.Discard()

	io.Copy(w, conn)

	dataID, err := w.Commit()
	if err != nil {
		return err
	}

	term.Printf("stored as %v (%s)\n", dataID, log.DataSize(dataID.Size))

	return nil
}

func (adm *Admin) info(term admin.Terminal, args []string) error {
	var f = "%-32s %6s %s\n"

	// list openers
	openers := adm.mod.openers.Values()
	slices.SortFunc(openers, func(a, b *Opener) int {
		return cmp.Compare(a.Priority, b.Priority) * -1
	})

	term.Printf("Openers:\n")
	term.Printf(f, admin.Header("Name"), admin.Header("Prio"), admin.Header("Type"))
	for _, opener := range openers {
		term.Printf(
			f,
			opener.Name,
			strconv.FormatInt(int64(opener.Priority), 10),
			reflect.TypeOf(opener.Opener),
		)
	}
	term.Println()

	// list creators
	creators := adm.mod.creators.Values()
	slices.SortFunc(creators, func(a, b *Creator) int {
		return cmp.Compare(a.Priority, b.Priority) * -1
	})

	term.Printf("Creators:\n")
	term.Printf(f, admin.Header("Prio"), admin.Header("Name"), admin.Header("Type"))
	for _, creator := range creators {
		term.Printf(
			f,
			creator.Name,
			strconv.FormatInt(int64(creator.Priority), 10),
			reflect.TypeOf(creator.Creator),
		)
	}
	term.Println()

	return nil
}

func (adm *Admin) ShortDescription() string {
	return "manage storage"
}

func (adm *Admin) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: storage <command>\n\n")
	term.Printf("commands:\n")
	term.Printf("  read [dataID]                             read data by ID (caution - may print binary data)\n")
	term.Printf("  fetch <url>                               download data to storage\n")
	term.Printf("  info                                      show info\n")
	term.Printf("  help                                      show help\n")
	return nil
}
