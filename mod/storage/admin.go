package storage

import (
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin/api"
	"io"
	"net/http"
	"time"
)

const defaultAccessDuration = time.Hour * 24 * 365 * 100 // 100 years

type Admin struct {
	mod  *Module
	cmds map[string]func(admin.Terminal, []string) error
}

func NewAdmin(mod *Module) *Admin {
	var adm = &Admin{mod: mod}
	adm.cmds = map[string]func(admin.Terminal, []string) error{
		"grant":  adm.grant,
		"revoke": adm.revoke,
		"access": adm.access,
		"read":   adm.read,
		"ls":     adm.ls,
		"get":    adm.get,
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

func (adm *Admin) grant(term admin.Terminal, args []string) error {
	if len(args) < 2 {
		return errors.New("argument missing")
	}

	var err error
	var identity id.Identity
	var dataID data.ID
	var expiresAt = time.Now().Add(defaultAccessDuration)

	if identity, err = adm.mod.node.Resolver().Resolve(args[0]); err != nil {
		return err
	}
	if dataID, err = data.Parse(args[1]); err != nil {
		return err
	}
	if len(args) >= 3 {
		d, err := time.ParseDuration(args[2])
		if err != nil {
			return err
		}
		expiresAt = time.Now().Add(d)
	}

	return adm.mod.Access().Grant(identity, dataID, expiresAt)
}

func (adm *Admin) revoke(term admin.Terminal, args []string) error {
	if len(args) < 2 {
		return errors.New("argument missing")
	}

	var err error
	var identity id.Identity
	var dataID data.ID

	if identity, err = adm.mod.node.Resolver().Resolve(args[0]); err != nil {
		return err
	}
	if dataID, err = data.Parse(args[1]); err != nil {
		return err
	}

	return adm.mod.Access().Revoke(identity, dataID)
}

func (adm *Admin) access(term admin.Terminal, args []string) error {
	var list []dbAccess

	tx := adm.mod.db.Limit(100).Find(&list)
	if tx.Error != nil {
		return tx.Error
	}
	f := "%-20s%-16s%s\n"

	term.Printf("showing %d results\n", len(list))
	term.Printf(f, admin.Header("IDENTITY"), admin.Header("EXPIRY"), admin.Header("DATAID"))

	for _, item := range list {
		access, err := item.toAccess()
		if err != nil {
			term.Printf("data error: %s\n", err)
			continue
		}

		var expiry = "expired"

		if access.ExpiresAt.After(time.Now()) {
			expiry = time.Until(access.ExpiresAt).Round(time.Second).String()
		}

		term.Printf(f,
			access.Identity,
			expiry,
			access.DataID)
	}

	return nil
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

		r, err := adm.mod.Data().Read(dataID, nil)
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

	term.Printf("downloading %v...\n", url)

	// Make a GET request to the URL
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	w, err := adm.mod.Data().Store(int(response.ContentLength))
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

func (adm *Admin) ls(term admin.Terminal, args []string) error {
	var since = time.Time{}

	var format = "%10s %-66s %s\n"
	var files = adm.mod.Data().IndexSince(since)
	var total int
	term.Printf(format, admin.Header("Size"), admin.Header("ID"), admin.Header("Indexed"))
	for _, file := range files {
		term.Printf(format, log.DataSize(file.ID.Size), file.ID, file.IndexedAt)
		total += int(file.ID.Size)
	}

	term.Printf("%d files, %v total\n", len(files), log.DataSize(total))

	return nil
}

func (adm *Admin) ShortDescription() string {
	return "manage storage"
}

func (adm *Admin) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: storage <command>\n\n")
	term.Printf("commands:\n")
	term.Printf("  grant <identity> <dataID> [duration]      grant access to data\n")
	term.Printf("  revoke <identity> <dataID>                revoke access to data\n")
	term.Printf("  access                                    show all access entries\n")
	term.Printf("  ls                                        list all indexed data\n")
	term.Printf("  read [dataID]                             read data by ID (caution - may print binary data)\n")
	term.Printf("  get <url>                                 download data over http(s)\n")
	term.Printf("  help                                      show help\n")
	return nil
}
