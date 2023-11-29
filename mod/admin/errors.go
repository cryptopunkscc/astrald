package admin

import "fmt"

type errModuleNotLoaded struct {
	Module string
}

func (e errModuleNotLoaded) Error() string {
	return fmt.Sprintf("module not loaded: %s", e.Module)
}
