package astral

import "fmt"

type ErrUnexpectedObject struct {
	Object Object
}

var _ error = &ErrUnexpectedObject{}

func NewErrUnexpectedObject(object Object) *ErrUnexpectedObject {
	return &ErrUnexpectedObject{Object: object}
}

func (e *ErrUnexpectedObject) Error() string {
	return fmt.Sprintf("unexpected object: %s", e.Object.ObjectType())
}

func (e *ErrUnexpectedObject) Is(other error) bool {
	_, ok := other.(*ErrUnexpectedObject)
	return ok
}
