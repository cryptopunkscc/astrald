package user

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/user"
	"time"
)

type Admin struct {
	mod  *Module
	cmds map[string]func(admin.Terminal, []string) error
}

func NewAdmin(mod *Module) *Admin {
	var adm = &Admin{mod: mod}
	adm.cmds = map[string]func(admin.Terminal, []string) error{
		"set":            adm.set,
		"nodes":          adm.nodes,
		"owner":          adm.owner,
		"claim":          adm.claim,
		"info":           adm.info,
		"contacts":       adm.contacts,
		"add_contact":    adm.addContact,
		"rm_contact":     adm.rmContact,
		"local_contract": adm.localContract,
		"help":           adm.help,
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

func (adm *Admin) nodes(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	userID, err := adm.mod.Dir.ResolveIdentity(args[0])
	if err != nil {
		return err
	}

	nodes := adm.mod.Nodes(userID)

	for _, node := range nodes {
		term.Printf("%v\n", node)
	}

	return nil
}

func (adm *Admin) addContact(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	userID, err := adm.mod.Dir.ResolveIdentity(args[0])
	if err != nil {
		return err
	}

	return adm.mod.AddContact(userID)
}

func (adm *Admin) rmContact(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	userID, err := adm.mod.Dir.ResolveIdentity(args[0])
	if err != nil {
		return err
	}

	return adm.mod.RemoveContact(userID)
}

func (adm *Admin) contacts(term admin.Terminal, args []string) error {
	for _, userID := range adm.mod.Contacts() {
		term.Printf("%v\n", userID)
	}
	return nil
}

func (adm *Admin) localContract(term admin.Terminal, args []string) error {
	c, err := adm.mod.LocalContract()
	if err != nil {
		return err
	}

	contractID, err := astral.ResolveObjectID(c)
	if err != nil {
		return err
	}

	j, _ := json.MarshalIndent(c, "", "  ")

	term.Printf("%v\n", contractID)
	term.Printf("%v\n", string(j))

	return nil
}

func (adm *Admin) owner(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	nodeID, err := adm.mod.Dir.ResolveIdentity(args[0])
	if err != nil {
		return err
	}

	userID := adm.mod.Owner(nodeID)

	if userID.IsZero() {
		return errors.New("user unknown")
	}
	term.Printf("%v\n", userID)

	return nil
}

func (adm *Admin) claim(term admin.Terminal, args []string) error {
	var d time.Duration

	if len(args) < 1 {
		return errors.New("missing argument")
	}

	nodeID, err := adm.mod.Dir.ResolveIdentity(args[0])
	if err != nil {
		return err
	}

	con, _ := adm.mod.Remote(nodeID, term.UserIdentity())

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	return con.Claim(ctx, d)
}

func (adm *Admin) set(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	identity, err := adm.mod.Dir.ResolveIdentity(args[0])
	if err != nil {
		return err
	}

	return adm.mod.SetUserID(identity)
}

func (adm *Admin) info(term admin.Terminal, args []string) error {
	userID := adm.mod.UserID()
	if userID.IsZero() {
		return errors.New("no user identity set")
	}

	term.Printf("Identity: %v\n", userID)
	term.Printf("PubKey:   %v\n", userID.String())

	contract, err := adm.mod.LocalContract()
	if err != nil {
		return err
	}
	contractID, err := astral.ResolveObjectID(contract)
	if err != nil {
		return err
	}
	term.Printf("Contract: %v\n", contractID)

	return nil
}

func (adm *Admin) ShortDescription() string {
	return "manage user"
}

func (adm *Admin) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: %v <command>\n\n", user.ModuleName)
	term.Printf("commands:\n")
	term.Printf("  set <identity>       set local user identity\n")
	term.Printf("  info                 show user info\n")
	term.Printf("  help                 show help\n")
	return nil
}
