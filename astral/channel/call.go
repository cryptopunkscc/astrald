package channel

import (
	"errors"
	"io"
	"reflect"

	"github.com/cryptopunkscc/astrald/astral"
)

var objectType = reflect.TypeOf((*astral.Object)(nil)).Elem()
var errorType = reflect.TypeOf((*error)(nil)).Elem()

func (ch Channel) Call(args ...any) error {
	var ctx *astral.Context

	if len(args) == 0 {
		return errors.New("no arguments provided")
	}

	var firstVal = reflect.ValueOf(args[0])
	if isContext(firstVal.Type()) {
		ctx = firstVal.Interface().(*astral.Context)
		args = args[1:]
		go func() {
			<-ctx.Done()
			ch.Close()
		}()
	}

	var funcMap = map[string]*funcWrapper{}
	var funcSet []*funcWrapper

	for _, arg := range args {
		pf, err := wrapFunc(arg)
		if err != nil {
			return err
		}
		if pf.ObjectType != "" {
			funcMap[pf.ObjectType] = pf
		} else {
			funcSet = append(funcSet, pf)
		}
	}

	for {
		obj, err := ch.Receive()
		switch {
		case err == nil:
		case errors.Is(err, io.EOF):
			return nil
		default:
			return err
		}

		handler, ok := funcMap[obj.ObjectType()]
		if !ok {
			for _, h := range funcSet {
				if h.canCall(obj) {
					handler = h
					break
				}
			}
		}

		if handler == nil {
			return astral.NewErrUnexpectedObject(obj)
		}

		err = handler.call(obj)
		switch err {
		case nil:
		case ErrStop:
			return nil
		default:
			return err
		}
	}
}

func ExpectAck(*astral.Ack) error {
	return ErrStop
}

func StopOnEOS(*astral.EOS) error {
	return ErrStop
}

func PassErrors(err error) error {
	return err
}

func Collect[T astral.Object](dst *[]T) func(T) error {
	return func(v T) error {
		*dst = append(*dst, v)
		return nil
	}
}

func isContext(t reflect.Type) bool {
	switch {
	case t.Kind() != reflect.Ptr:
		return false
	case t.Elem() == reflect.TypeOf((*astral.Context)(nil)).Elem():
		return true
	default:
		return false
	}
}
