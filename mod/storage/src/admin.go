package storage

import (
	"errors"
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
	"strings"
)

type Admin struct {
	mod  *Module
	cmds map[string]func(admin.Terminal, []string) error
}

func NewAdmin(mod *Module) *Admin {
	var adm = &Admin{mod: mod}
	adm.cmds = map[string]func(admin.Terminal, []string) error{
		"read": adm.read,
		"get":  adm.get,
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

func (adm *Admin) read(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	for _, idstr := range args {
		dataID, err := data.Parse(idstr)
		if err != nil {
			return err
		}

		r, err := adm.mod.Read(dataID, nil)
		if err != nil {
			return err
		}

		io.Copy(term, r)
	}

	return nil
}

func (adm *Admin) get(term admin.Terminal, args []string) error {
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

		w, err := adm.mod.Store(
			&storage.StoreOpts{
				Alloc: int(response.ContentLength),
			},
		)
		if err != nil {
			return err
		}
		defer w.Discard()

		io.Copy(w, response.Body)

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

	w, err := adm.mod.Store(nil)
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
	var f = "%-32s %s\n"
	var names []string

	// list readers
	names = adm.mod.readers.Keys()
	slices.Sort(names)

	term.Printf(f, admin.Header("Reader"), admin.Header("Type"))
	for _, name := range names {
		v, ok := adm.mod.readers.Get(name)
		if !ok {
			continue
		}
		term.Printf(f, name, reflect.TypeOf(v))
	}
	term.Println()

	// list stores
	names = adm.mod.stores.Keys()
	slices.Sort(names)

	term.Printf(f, admin.Header("Store"), admin.Header("Type"))
	for _, name := range names {
		v, ok := adm.mod.stores.Get(name)
		if !ok {
			continue
		}
		term.Printf(f, name, reflect.TypeOf(v))
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
	term.Printf("  get <url>                                 download data over http(s)\n")
	term.Printf("  info                                      show info\n")
	term.Printf("  help                                      show help\n")
	return nil
}
