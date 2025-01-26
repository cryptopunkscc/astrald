package apphost

import (
	"errors"
	"github.com/cryptopunkscc/astrald/mod/admin"
)

type Admin struct {
	mod *Module
}

func (adm *Admin) Exec(t admin.Terminal, args []string) error {
	if len(args) <= 1 {
		return adm.help(t)
	}

	switch args[1] {
	case "tokens":
		return adm.tokens(t, args[2:])

	case "newtoken":
		return adm.newtoken(t, args[2:])

	case "help":
		return adm.help(t)

	default:
		return errors.New("unknown command")
	}
}

func (adm *Admin) newtoken(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing: identity")
	}

	identity, err := adm.mod.Dir.ResolveIdentity(args[0])
	if err != nil {
		return err
	}

	token, err := adm.mod.CreateAccessToken(identity)
	if err != nil {
		return err
	}

	term.Printf("New access token: %v\n", token)

	return nil
}

func (adm *Admin) tokens(term admin.Terminal, args []string) error {
	var rows []dbAccessToken

	adm.mod.db.Find(&rows)

	const f = "%v %v\n"

	term.Printf(f, admin.Header("Token"), admin.Header("Identity"))

	for _, row := range rows {
		term.Printf(f, admin.Keyword(row.Token), row.Identity)
	}

	return nil
}

func (adm *Admin) help(out admin.Terminal) error {
	out.Println("usage: apphost <command>")
	out.Println()
	out.Println("commands:")
	out.Println("  tokens                  list all access tokens")
	out.Println("  newtoken <identity>     create new access token for an identity")
	out.Println("  help                    show help")

	return nil
}

func (adm *Admin) ShortDescription() string {
	return "manage application host"
}
