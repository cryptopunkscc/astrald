package astral

import (
	"fmt"
	"io"
)

type ErrUnexpectedObject struct {
	Object Object
}

var _ JSONObject = &ErrUnexpectedObject{}
var _ error = &ErrUnexpectedObject{}

func NewErrUnexpectedObject(object Object) *ErrUnexpectedObject {
	return &ErrUnexpectedObject{Object: object}
}

func (ErrUnexpectedObject) ObjectType() string {
	return "err_unexpected_object"
}

func (e ErrUnexpectedObject) WriteTo(writer io.Writer) (n int64, err error) {
	return Objectify(&e).WriteTo(writer)
}

func (e *ErrUnexpectedObject) ReadFrom(reader io.Reader) (n int64, err error) {
	return Objectify(e).ReadFrom(reader)
}

func (e ErrUnexpectedObject) MarshalJSON() ([]byte, error) {
	return Objectify(&e).MarshalJSON()
}

func (e *ErrUnexpectedObject) UnmarshalJSON(bytes []byte) error {
	return Objectify(e).UnmarshalJSON(bytes)
}

func (e *ErrUnexpectedObject) Error() string {
	return fmt.Sprintf("unexpected object: %s", e.Object.ObjectType())
}

func (e *ErrUnexpectedObject) Is(other error) bool {
	_, ok := other.(*ErrUnexpectedObject)
	return ok
}

func init() {
	_ = Add(&ErrUnexpectedObject{})
}
