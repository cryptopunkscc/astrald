package shell

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/lib/term"
	"github.com/cryptopunkscc/astrald/sig"
	"io"
	"net/url"
	"reflect"
	"strings"
)

// Scope is a collection of operations and subscopes.
//
// Add an operation to the scope:
//
//	scope.AddOp("hi", func(ctx *astral.Context, query shell.Query) error {
//		conn, _ := query.Accept()
//	    return conn.Printf("hello, %v!\n", ctx.Identity())
//	})
type Scope struct {
	ops  sig.Map[string, any]
	subs sig.Map[string, *Scope]
	Log  *log.Logger
}

type HasScope interface {
	Scope() *Scope
}

func NewScope(log *log.Logger) *Scope {
	return &Scope{Log: log}
}

func (scope *Scope) AddOp(name string, op any) error {
	v := reflect.ValueOf(op)
	typ := v.Type()

	if v.Kind() != reflect.Func {
		return errors.New("op must be an op function")
	}
	if typ.NumIn() != 2 && typ.NumIn() != 3 {
		return errors.New("op must be an op function")
	}
	if !(typ.In(0) == reflect.TypeOf(&astral.Context{})) {
		return errors.New("op must be an op function")
	}

	if !typ.In(1).Implements(reflect.TypeOf((*Query)(nil)).Elem()) {
		return errors.New("op must be an op function")
	}

	if typ.NumIn() == 3 {
		switch typ.In(2).Kind() {
		case reflect.Pointer:
			if typ.In(2).Elem().Kind() != reflect.Struct {
				return errors.New("op must be an op function")
			}
		case reflect.Struct:
		default:
			return errors.New("op must be an op function")
		}
	}

	if typ.NumOut() != 1 {
		return errors.New("op must be an op function")
	}
	if !typ.Out(0).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return errors.New("op must be an op function")
	}

	_, ok := scope.ops.Set(name, op)
	if !ok {
		return errors.New("op already defined")
	}

	return nil
}

func (scope *Scope) AddStruct(s any, prefix string) (err error) {
	var errs []error
	v := reflect.ValueOf(s)

	if (v.Kind() != reflect.Pointer) || (v.Elem().Kind() != reflect.Struct) {
		return errors.New("argument must be a pointer to a struct")
	}

	for i := 0; i < v.NumMethod(); i++ {
		if !v.Method(i).CanInterface() {
			continue
		}

		m := v.Method(i).Interface()

		n, ok := strings.CutPrefix(v.Type().Method(i).Name, prefix)
		if !ok {
			continue
		}
		n = term.ToSnakeCase(n)

		if e := scope.AddOp(n, m); e != nil {
			errs = append(errs, e)
		}
	}

	return errors.Join(errs...)
}

func (scope *Scope) AddScope(name string, s *Scope) error {
	_, ok := scope.subs.Set(name, s)
	if !ok {
		return errors.New("scope already defined")
	}
	return nil
}

func (scope *Scope) Call(ctx *astral.Context, q Query, name string, args map[string]string) (err error) {
	var op = scope.getOp(name)
	if op == nil {
		return errors.New("op not found")
	}

	var fnArgs = []reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(q),
	}

	var fn = reflect.ValueOf(op)

	if fn.Type().NumIn() == 3 {
		var argType = fn.Type().In(2)
		var argVal reflect.Value

		switch argType.Kind() {
		case reflect.Ptr:
			argVal = reflect.New(argType.Elem())
			err = query.Populate(args, argVal.Interface())

		case reflect.Struct:
			argVal = reflect.New(argType)
			err = query.Populate(args, argVal.Interface())
			argVal = argVal.Elem()

		default:
			panic("invalid arg type")
		}

		if err != nil {
			return err
		}

		fnArgs = append(fnArgs, argVal)
	}

	var ret = fn.Call(fnArgs)[0]

	if ret.IsNil() {
		return nil
	}

	return ret.Interface().(error)
}

func (scope *Scope) CallQuery(ctx *astral.Context, q Query, name string, query string) (err error) {
	return scope.Call(ctx, q, name, ParseQuery(query))
}

func (scope *Scope) CallArgs(ctx *astral.Context, q Query, name string, args []string) (err error) {
	return scope.Call(ctx, q, name, ParseArgs(args))
}

func (scope *Scope) Ops() []string {
	return scope.ops.Keys()
}

func (scope *Scope) Subs() []string {
	return scope.subs.Keys()
}

func (scope *Scope) Tree() (tree []string) {
	for subName, sub := range scope.subs.Clone() {
		for _, n := range sub.Tree() {
			tree = append(tree, subName+"."+n)
		}
	}

	tree = append(tree, scope.ops.Keys()...)

	return
}

func (scope *Scope) Exists(name string) (found bool) {
	return scope.getOp(name) != nil
}

func (scope *Scope) getOp(name string) any {
	if idx := strings.IndexByte(name, '.'); idx != -1 {
		if sub, ok := scope.subs.Get(name[:idx]); ok {
			return sub.getOp(name[idx+1:])
		}
	}
	op, _ := scope.ops.Get(name)
	return op
}

func (scope *Scope) RouteQuery(ctx *astral.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	path, params := query.Parse(q.Query)

	if !scope.Exists(path) {
		return query.RouteNotFound(scope)
	}

	var query = NewNetworkQuery(w, q)
	defer query.Reject()

	go func() {
		ctx := astral.NewContext(nil).WithIdentity(ctx.Identity())
		err := scope.Call(ctx, query, path, params)
		if err != nil {
			scope.Log.Errorv(1, "failed to call query %v: %v", path, err)
			query.Reject()
		}
	}()

	return query.Resolve()
}

func ParseQuery(q string) (params map[string]string) {
	vals, err := url.ParseQuery(q)
	if err != nil {
		return
	}

	params = make(map[string]string)
	for k, v := range vals {
		if len(v) > 0 {
			params[k] = v[0]
		} else {
			params[query.DefaultArgKey] = k
		}
	}

	return
}

func ParseArgs(args []string) (params map[string]string) {
	params = make(map[string]string)

	for len(args) > 0 {
		key := args[0]

		if !strings.HasPrefix(key, "-") {
			params[query.DefaultArgKey] = key
			args = args[1:]
			continue
		}

		key = key[1:]

		if len(args) < 2 {
			params[key] = ""
			return
		}

		params[key] = args[1]
		args = args[2:]
	}
	return
}
