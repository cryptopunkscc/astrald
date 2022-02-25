package cslq

import (
	"errors"
	"fmt"
	"reflect"
)

type ErrUnexpectedToken struct {
	Token interface{}
}

func (e ErrUnexpectedToken) Error() string {
	return fmt.Sprintf("unexpected token %s", reflect.TypeOf(e.Token))
}

type ErrInvalidOp struct {
	Op interface{}
}

func (e ErrInvalidOp) Error() string {
	return fmt.Sprintf("invalid op %s", reflect.TypeOf(e.Op))
}

type ErrInvalidDataLength struct {
	Expected int
	Actual   int
}

func (e ErrInvalidDataLength) Error() string {
	return fmt.Sprintf("invalid data length, expected %d, got %d", e.Expected, e.Actual)
}

type ErrCannotConvert struct {
	From string
	To   string
}

func (e ErrCannotConvert) Error() string {
	return fmt.Sprintf("cannot convert from %s to %s", e.From, e.To)
}

var ErrNotAPointer = errors.New("variable is not a pointer")
var ErrNotAStructPointer = errors.New("variable is not a struct pointer")
var ErrCannotDecodeString = errors.New("cannot decode array into string")
var ErrNotEnoughArguments = errors.New("not enough arguments")
