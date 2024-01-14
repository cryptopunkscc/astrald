package modules

import (
	"fmt"
)

type ErrModuleUnavailable struct {
	Name string
}

func (err ErrModuleUnavailable) Error() string {
	return fmt.Sprintf("module %s unavailable", err.Name)
}

func ModuleUnavailable(name string) ErrModuleUnavailable {
	return ErrModuleUnavailable{Name: name}
}

func (ErrModuleUnavailable) Is(other error) bool {
	_, ok := other.(*ErrModuleUnavailable)
	return ok
}
