package ops

import (
	"errors"
	"reflect"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
)

/*
Op wraps an op function of this signature:

 	type Op func(ctx *astral.Context, q *Query) error

With an optional third argument that can be a struct with fields representing operation arguments:

	type ArgType struct {
		ID      *astral.Identity
        Message string
	}

	type Op func(ctx *astral.Context, q *Query, args ArgType) error

Then allows this function to be invoked via the Call() function:

	op.Call(ctx, query, map[string]string{
		"id": "033e9ce4e8f0e68f7db49ffb6b9eecc10605f3f3fcb3c630545887749ab515b9c7",
		"message": "Hello, world!",
    }) error
*/

type Op struct {
	fn       reflect.Value                                  // the function to be called
	argType  reflect.Type                                   // optional type of the 3rd func argument
	hasArgs  bool                                           // whether the function has a 3rd argument
	populate func(map[string]string) (reflect.Value, error) // populates the map into the args
}

var queryType = reflect.TypeOf((*Query)(nil)).Elem()

// Func wraps a function as an Op. fn has to have one of the following signatures:
// - func(*astral.Context, *shell.Query) error
// - func(*astral.Context, *shell.Query, ArgType) error
// ArgType can ba struct or a pointer to a struct with all fields exported.
func Func(fn any) (*Op, error) {
	var v = reflect.ValueOf(fn)
	var t = v.Type()
	var c = &Op{fn: v} // validate the fn argument

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
		c.argType = v.Type().In(2) // argument type
		c.hasArgs = true

		switch c.argType.Kind() {
		case reflect.Pointer:
			if c.argType.Elem().Kind() != reflect.Struct {
				return nil, errors.New("third argument is a non-struct pointer")
			}
			c.populate = func(m map[string]string) (av reflect.Value, err error) {
				av = reflect.New(c.argType.Elem())
				err = query.Populate(m, av.Interface())
				return
			}

		case reflect.Struct:
			c.populate = func(m map[string]string) (av reflect.Value, err error) {
				av = reflect.New(c.argType)
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

	if c.fn.Type().NumIn() == 3 {
		a, err := c.populate(args)
		if err != nil {
			return err
		}

		fnArgs = append(fnArgs, a)
	}

	var ret = c.fn.Call(fnArgs)[0]

	if ret.IsNil() {
		return nil
	}

	return ret.Interface().(error)
}

func (c *Op) ArgumentSpecs() (args []query.FieldSpec) {
	if !c.hasArgs {
		return
	}

	wrapper, err := query.Edit(reflect.New(c.argType).Interface())
	if err != nil {
		panic(err)
	}

	return wrapper.Spec()
}
