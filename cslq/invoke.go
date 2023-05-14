package cslq

import (
	"io"
)

func Invokef[T any](r io.Reader, format string, fn func(T) error) error {
	var v T
	if err := Decode(r, format, &v); err != nil {
		return err
	}
	return fn(v)
}

func Invoke[T any](r io.Reader, fn func(T) error) error {
	return Invokef(r, "v", fn)
}
