package admin

import . "github.com/cryptopunkscc/astrald/mod/admin/api"

func (mod *Module) AddCommand(name string, cmd Command) error {
	mod.mu.Lock()
	defer mod.mu.Unlock()

	mod.commands[name] = cmd
	return nil
}
