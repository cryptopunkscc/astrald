package admin

import (
	"errors"
	"github.com/cryptopunkscc/astrald/mod/admin"
)

var _ admin.Command = &CmdUse{}

type CmdUse struct {
	mod *Module
}

func (cmd *CmdUse) Exec(term admin.Terminal, args []string) error {
	if len(args) < 2 {
		return cmd.help(term, nil)
	}

	enterCmd := args[1]

	if _, found := cmd.mod.commands[enterCmd]; !found {
		return errors.New("command not found")
	}

	term.Printf("type exit to go back\n")

	for {
		term.Printf("%v@%v:%v%v", term.UserIdentity(), cmd.mod.node.Identity(), admin.Keyword(enterCmd), cmd.mod.config.Prompt)

		line, err := term.ScanLine()
		if err != nil {
			return err
		}

		if line == "exit" {
			return nil
		}

		if err := cmd.mod.exec(enterCmd+" "+line, term); err != nil {
			term.Printf("error: %v\n", err)
		} else {
			term.Printf("ok\n")
		}
	}
}

func (cmd *CmdUse) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: use <command>\n")
	return nil
}

func (cmd *CmdUse) ShortDescription() string {
	return "enter the context of a command"
}
