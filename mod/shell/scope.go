package shell

import (
	"errors"
	"io"
	"reflect"
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/sig"
)

// Scope is a collection of operations and subscopes.
//
// Add an operation to the scope:
//
//	scope.AddFunc("hi", func(ctx *astral.Context, query shell.Query) error {
//		conn, _ := query.Accept()
//	    return conn.Printf("hello, %v!\n", ctx.Identity())
//	})
type Scope struct {
	ops  sig.Map[string, *Op]
	subs sig.Map[string, *Scope]
	Log  *log.Logger
}

type HasScope interface {
	Scope() *Scope
}

func NewScope(log *log.Logger) *Scope {
	return &Scope{Log: log}
}

func (scope *Scope) AddFunc(name string, fn any) error {
	op, err := Func(fn)
	if err != nil {
		return err
	}
	scope.ops.Set(name, op)
	return nil
}

// AddStruct adds all methods of a struct to the scope that start with the given prefix.
func (scope *Scope) AddStruct(s any, prefix string) (err error) {
	var errs []error
	v := reflect.ValueOf(s)

	if (v.Kind() != reflect.Pointer) || (v.Elem().Kind() != reflect.Struct) {
		return errors.New("argument must be a pointer to a struct")
	}

	for i := range v.NumMethod() {
		// skip unexported methods
		if !v.Method(i).CanInterface() {
			continue
		}

		fn := v.Method(i).Interface()

		name, hadPrefix := strings.CutPrefix(v.Type().Method(i).Name, prefix)
		if !hadPrefix {
			continue // skip methods without the prefix
		}

		name = log.ToSnakeCase(name)

		if e := scope.AddFunc(name, fn); e != nil {
			errs = append(errs, e)
		}
	}

	return errors.Join(errs...)
}

// AddScope adds a subscope to the scope.
func (scope *Scope) AddScope(name string, s *Scope) error {
	_, ok := scope.subs.Set(name, s)
	if !ok {
		return errors.New("scope already defined")
	}
	return nil
}

func (scope *Scope) Ops() []string {
	return scope.ops.Keys()
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

func (scope *Scope) Find(name string) (op *Op) {
	if idx := strings.IndexByte(name, '.'); idx != -1 {
		if sub, ok := scope.subs.Get(name[:idx]); ok {
			return sub.Find(name[idx+1:])
		}
	}
	op, _ = scope.ops.Get(name)
	return
}

func (scope *Scope) RouteQuery(ctx *astral.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	path, params := query.Parse(q.Query)

	op := scope.Find(path)
	if op == nil {
		return query.RouteNotFound(scope)
	}

	var query = NewNetworkQuery(w, q)
	defer query.Reject()

	go func() {
		// ctx will end as soon as the query resolves, so we need a new one for the op
		ctx := astral.NewContext(nil).WithIdentity(ctx.Identity())

		// call the op
		err := op.Call(ctx, query, params)
		if err != nil && scope.Log != nil {
			scope.Log.Errorv(1, "call %v: %v", path, err)
		}

		// reject the query in case the op did not respond to it, will do nothing if it did.
		_ = query.Reject()
	}()

	return query.Resolve(ctx)
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
