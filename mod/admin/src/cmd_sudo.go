package admin

import (
	"errors"
	"github.com/cryptopunkscc/astrald/mod/admin"
)

var _ admin.Command = &CmdSudo{}

type CmdSudo struct {
	mod *Module
}

func (cmd *CmdSudo) Exec(term admin.Terminal, args []string) error {
	if len(args) < 2 {
		return cmd.help(term, nil)
	}

	targetID, err := cmd.mod.Dir.ResolveIdentity(args[1])
	if err != nil {
		return err
	}

	if !cmd.mod.Auth.Authorize(term.UserIdentity(), admin.ActionSudo, targetID) {
		return errors.New("unauthorized")
	}

	term.SetUserIdentity(targetID)

	return nil
}

func (cmd *CmdSudo) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: sudo <identity>\n")
	return nil
}

func (cmd *CmdSudo) ShortDescription() string {
	return "switch user identity"
}
