package ops

import (
	"errors"
	"reflect"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/lib/query"
)

type Op struct {
	v        reflect.Value
	at       reflect.Type
	hasArgs  bool
	populate func(map[string]string) (reflect.Value, error)
}

var queryType = reflect.TypeOf((*Query)(nil)).Elem()

// Func wraps a function as an Op. fn has to have one of the following signatures:
// - func(*astral.Context, *shell.Query) error
// - func(*astral.Context, *shell.Query, ArgType) error
// ArgType can ba struct or a pointer to a struct with all fields exported.
func Func(fn any) (*Op, error) {
	var v = reflect.ValueOf(fn)
	var t = v.Type()
	var c = &Op{v: v} // validate the fn argument

	switch {
	case v.Kind() != reflect.Func:
		return nil, errors.New("fn must be a function")
	case t.NumIn() != 2 && t.NumIn() != 3:
		return nil, errors.New("invalid number of arguments")
	case t.In(0) != reflect.TypeOf(&astral.Context{}):
		return nil, errors.New("first argument must be a *astral.Context")
	case t.In(1).Kind() != reflect.Ptr:
		return nil, errors.New("second argument must be an *ops.Query")
	case t.In(1).Elem() != queryType:
		return nil, errors.New("second argument must be an *ops.Query")
	case t.NumOut() != 1 || !t.Out(0).Implements(reflect.TypeOf((*error)(nil)).Elem()):
		return nil, errors.New("fn must return a single error value")
	case t.NumIn() == 3:
		c.at = v.Type().In(2) // argument type
		c.hasArgs = true

		switch c.at.Kind() {
		case reflect.Pointer:
			if c.at.Elem().Kind() != reflect.Struct {
				return nil, errors.New("third argument is a non-struct pointer")
			}
			c.populate = func(m map[string]string) (av reflect.Value, err error) {
				av = reflect.New(c.at.Elem())
				err = query.Populate(m, av.Interface())
				return
			}

		case reflect.Struct:
			c.populate = func(m map[string]string) (av reflect.Value, err error) {
				av = reflect.New(c.at)
				err = query.Populate(m, av.Interface())
				av = av.Elem()
				return
			}

		default:
			return nil, errors.New("third argument must be a struct")
		}
	}

	return c, nil
}

// Call calls the Op
func (c *Op) Call(ctx *astral.Context, q *Query, args map[string]string) error {
	var fnArgs = []reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(q),
	}

	if c.v.Type().NumIn() == 3 {
		a, err := c.populate(args)
		if err != nil {
			return err
		}

		fnArgs = append(fnArgs, a)
	}

	var ret = c.v.Call(fnArgs)[0]

	if ret.IsNil() {
		return nil
	}

	return ret.Interface().(error)
}

func (c *Op) ArgNames() (names []string) {
	if !c.hasArgs {
		return
	}

	for i := range c.at.NumField() {
		ft := c.at.Field(i)
		names = append(names, log.ToSnakeCase(ft.Name))
	}

	return
}
