package astral

import (
	"encoding/json"
	"io"
)

type Error interface {
	Object
	error
}

var _ Error = &ErrorMessage{}

type ErrorMessage struct {
	err string
}

func NewError(s string) Error {
	return &ErrorMessage{
		err: s,
	}
}

func Err(err error) Error {
	return NewError(err.Error())
}

// astral

func (msg ErrorMessage) ObjectType() string {
	return "error_message"
}

func (msg ErrorMessage) WriteTo(w io.Writer) (n int64, err error) {
	return String16(msg.err).WriteTo(w)
}

func (msg *ErrorMessage) ReadFrom(r io.Reader) (n int64, err error) {
	return (*String16)(&msg.err).ReadFrom(r)
}

// json

func (msg ErrorMessage) MarshalJSON() ([]byte, error) {
	return json.Marshal(msg.err)
}

func (msg *ErrorMessage) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, &msg)
}

// text

func (msg ErrorMessage) MarshalText() (text []byte, err error) {
	return []byte(msg.err), nil
}

func (msg *ErrorMessage) UnmarshalText(text []byte) error {
	msg.err = string(text)
	return nil
}

// other

func (msg ErrorMessage) Error() string {
	return msg.err
}

func (msg ErrorMessage) String() string {
	return msg.err
}

func init() {
	_ = Add(&ErrorMessage{})
}
