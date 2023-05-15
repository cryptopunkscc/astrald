package admin

type Command interface {
	Exec(out *Terminal, args []string) error
}

func (mod *Module) AddCommand(name string, cmd Command) error {
	mod.commands[name] = cmd
	return nil
}
