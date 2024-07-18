package relay

import (
	"errors"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/relay"
	"time"
)

type Admin struct {
	mod  *Module
	cmds map[string]func(admin.Terminal, []string) error
}

func NewAdmin(mod *Module) *Admin {
	var adm = &Admin{mod: mod}
	adm.cmds = map[string]func(admin.Terminal, []string) error{
		"certs":  adm.certs,
		"mkcert": adm.mkcert,
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

func (adm *Admin) certs(term admin.Terminal, args []string) error {
	objectIDs, err := adm.mod.FindCerts(&relay.FindOpts{IncludeExpired: true})
	if err != nil {
		return err
	}

	const f = "%-64s %-33s %-33s %-10s %-20s\n"

	term.Printf(f,
		admin.Header("ObjectID"),
		admin.Header("Target"),
		admin.Header("Relay"),
		admin.Header("Direction"),
		admin.Header("Expires at"),
	)

	for _, objectID := range objectIDs {
		object, err := adm.mod.objects.Get(objectID, nil)
		if err != nil {
			term.Printf("%v error: %v\n", objectID, err)
			continue
		}

		cert, err := relay.UnmarshalCert(object)
		if err != nil {
			term.Printf("%v error: %v\n", objectID, err)
			continue
		}

		term.Printf(
			f,
			objectID,
			cert.TargetID,
			cert.RelayID,
			admin.Keyword(cert.Direction),
			cert.ExpiresAt,
		)
	}

	return nil
}

func (adm *Admin) mkcert(term admin.Terminal, args []string) error {
	var targetID id.Identity
	var relayID = adm.mod.node.Identity()
	var duration = 30 * 24 * time.Hour
	var direction = relay.Both

	if len(args) < 1 {
		return errors.New("missing arguments")
	}

	targetID, err := adm.mod.dir.Resolve(args[0])
	if err != nil {
		return err
	}

	if len(args) >= 2 {
		relayID, err = adm.mod.dir.Resolve(args[1])
		if err != nil {
			return err
		}
	}

	certID, err := adm.mod.MakeCert(targetID, relayID, direction, duration)
	if err != nil {
		return err
	}

	term.Printf("Done! New certificate ID: %v\n", certID)

	return nil
}

func (adm *Admin) ShortDescription() string {
	return "manage the relay module"
}

func (adm *Admin) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: %s <command>\n\n", relay.ModuleName)
	term.Printf("commands:\n")
	term.Printf("  certs                       show all relay certificates\n")
	term.Printf("  mkcert <target> [relay]     create (and sign) a new certificate\n")
	term.Printf("  help                        show help\n")
	return nil
}
