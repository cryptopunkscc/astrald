package astral

import "fmt"

type ErrBlueprintNotFound struct {
	Type string
}

var _ error = &ErrBlueprintNotFound{}

func NewErrBlueprintNotFound(t string) error {
	return &ErrBlueprintNotFound{Type: t}
}

func (e *ErrBlueprintNotFound) Error() string {
	return fmt.Sprintf("blueprint not found: %s", e.Type)
}

func (e *ErrBlueprintNotFound) Is(other error) bool {
	_, ok := other.(*ErrBlueprintNotFound)
	return ok
}
