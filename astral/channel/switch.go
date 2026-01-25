package channel

import (
	"errors"
	"io"
	"reflect"

	"github.com/cryptopunkscc/astrald/astral"
)

var objectType = reflect.TypeOf((*astral.Object)(nil)).Elem()
var errorType = reflect.TypeOf((*error)(nil)).Elem()
var configType = reflect.TypeOf((*Config)(nil)).Elem()

// Switch takes a list of functions with a single argument and an optional return value of an error type.
// It then receives objects from the channel and passes them to functions with compatible argument types.
// It stops on Receive() error or when a function returns an error. If a function returns ErrStop, Switch
// stops end returns nil. If no function takes the object type, an ErrUnexpectedObject is returned.
//
// Example:
//
//	 	ch := channel.New(channel.Join(os.Stdin, os.Stdout))
//		ch.Switch(
//			func(s *astral.String8) {
//				fmt.Println(*s)
//			},
//			channel.StopOnEOS,
//		)
//
// Switch supports config functions WithTimeout() and WithContext().
func (ch Channel) Switch(args ...any) error {
	if len(args) == 0 {
		return errors.New("no arguments provided")
	}

	var funcMap = map[string]*funcWrapper{}
	var funcSet []*funcWrapper
	var configSet []ConfigFunc

	for _, arg := range args {
		// catch ConfigFunc first
		if configFunc := toConfigFunc(arg); configFunc != nil {
			configSet = append(configSet, configFunc)
			continue
		}

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

	var cfg Config
	for _, f := range configSet {
		f(&cfg)
	}

	done := make(chan struct{})
	defer close(done)

	if cfg.cancelCh != nil {
		go func() {
			select {
			case <-done:
			case <-cfg.cancelCh:
				ch.Close()
			}
		}()
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

// ExpectAck stops the Switch function on Ack.
func ExpectAck(*astral.Ack) error {
	return ErrStop
}

// StopOnEOS stops the Switch function on EOS.
func StopOnEOS(*astral.EOS) error {
	return ErrStop
}

func PassErrors(err error) error {
	return err
}

// Collect appends objects to a slice. Example:
//
//	var list []*astral.String8
//	ch.Switch(channel.Collect(&list), channel.StopOnEOS)
func Collect[T astral.Object](dst *[]T) func(T) error {
	return func(v T) error {
		*dst = append(*dst, v)
		return nil
	}
}

// Chan sends objects to a go channel. Example:
//
//	var ch := make(chan *astral.String8)
//	ch.Switch(channel.Chan(ch), channel.StopOnEOS)
func Chan[T astral.Object](dst chan<- T) func(T) error {
	return func(v T) error {
		dst <- v
		return nil
	}
}

func toConfigFunc(v any) ConfigFunc {
	var f = reflect.ValueOf(v)
	switch {
	case f.Kind() != reflect.Func:
		return nil
	case f.Type().NumIn() != 1:
		return nil
	case f.Type().NumOut() != 0:
		return nil
	case f.Type().In(0).Kind() != reflect.Ptr:
		return nil
	case f.Type().In(0).Elem() == configType:
		return func(config *Config) {
			f.Call([]reflect.Value{reflect.ValueOf(config)})
		}
	}
	return nil
}
