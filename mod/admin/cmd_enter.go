package admin

import "errors"

var _ Command = &CmdEnter{}

type CmdEnter struct {
	mod *Module
}

func (cmd *CmdEnter) Exec(term *Terminal, args []string) error {
	if len(args) < 2 {
		return cmd.help(term, nil)
	}

	enterCmd := args[1]

	if _, found := cmd.mod.commands[enterCmd]; !found {
		return errors.New("command not found")
	}

	term.Printf("type exit to go back\n")

	for {
		term.Printf("%s[%s]%s", cmd.mod.node.Alias(), enterCmd, cmd.mod.config.Prompt)

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

func (cmd *CmdEnter) help(term *Terminal, _ []string) error {
	term.Printf("usage: enter <command>\n")
	return nil
}

func (cmd *CmdEnter) ShortDescription() string {
	return "enter a command"
}
