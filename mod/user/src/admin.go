package user

import (
	"encoding/json"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/user"
)

type Admin struct {
	mod  *Module
	cmds map[string]func(admin.Terminal, []string) error
}

func NewAdmin(mod *Module) *Admin {
	var adm = &Admin{mod: mod}
	adm.cmds = map[string]func(admin.Terminal, []string) error{
		"nodes":           adm.nodes,
		"active":          adm.active,
		"active_contract": adm.activeContract,
		"help":            adm.help,
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

	nodes := adm.mod.ActiveNodes(userID)

	for _, node := range nodes {
		term.Printf("%v\n", node)
	}

	return nil
}

func (adm *Admin) activeContract(term admin.Terminal, args []string) error {
	contract := adm.mod.ActiveContract()
	if contract == nil {
		term.Printf("no active contract\n")
		return nil
	}

	contractID, err := astral.ResolveObjectID(contract)
	if err != nil {
		return err
	}

	j, _ := json.MarshalIndent(contract, "", "  ")

	term.Printf("%v\n", contractID)
	term.Printf("%v\n", string(j))

	return nil
}

func (adm *Admin) active(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	nodeID, err := adm.mod.Dir.ResolveIdentity(args[0])
	if err != nil {
		return err
	}

	term.Printf("users on %v:\n", nodeID)
	for _, userID := range adm.mod.ActiveUsers(nodeID) {
		term.Printf("%v\n", userID)
	}

	return nil
}

func (adm *Admin) ShortDescription() string {
	return "manage user"
}

func (adm *Admin) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: %v <command>\n\n", user.ModuleName)
	term.Printf("commands:\n")
	term.Printf("  set <identity>       set local user identity\n")
	term.Printf("  help                 show help\n")
	return nil
}
