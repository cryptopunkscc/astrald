package ops

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

// Set is an astral.Router that routes queries to a set of operations.
//
// Add an operation to the set:
//
//	scope.AddFunc("hi", func(ctx *astral.Context, query shell.Query) error {
//		conn, _ := query.Accept()
//	    return conn.Printf("hello, %v!\n", ctx.Identity())
//	})
type Set struct {
	ops     sig.Map[string, *Op]
	subs    sig.Map[string, *Set]
	Log     *log.Logger
	OnError func(error, *astral.Query)
}

var _ astral.Router = &Set{}

type HasOps interface {
	GetOpSet() *Set
}

func NewSet() *Set {
	return &Set{}
}

func Struct(s any, prefix string) *Set {
	set := NewSet()
	err := set.AddStruct(s, prefix)
	if err != nil {
		panic(err)
	}
	return set
}

func (set *Set) AddFunc(name string, fn any) error {
	op, err := Func(fn)
	if err != nil {
		return err
	}
	set.ops.Set(name, op)
	return nil
}

// AddStruct adds to the set all methods of a struct that start with the given prefix (the prefix is removed).
func (set *Set) AddStruct(s any, prefix string) (err error) {
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

		if e := set.AddFunc(name, fn); e != nil {
			errs = append(errs, e)
		}
	}

	return errors.Join(errs...)
}

// AddSet adds another set as a subset.
func (set *Set) AddSet(name string, s *Set) error {
	_, ok := set.subs.Set(name, s)
	if !ok {
		return errors.New("set already defined")
	}
	return nil
}

func (set *Set) Tree() (tree []string) {
	for subName, sub := range set.subs.Clone() {
		for _, n := range sub.Tree() {
			tree = append(tree, subName+"."+n)
		}
	}

	tree = append(tree, set.ops.Keys()...)

	return
}

func (set *Set) Find(name string) (op *Op) {
	if idx := strings.IndexByte(name, '.'); idx != -1 {
		if sub, ok := set.subs.Get(name[:idx]); ok {
			return sub.Find(name[idx+1:])
		}
	}
	op, _ = set.ops.Get(name)
	return
}

func (set *Set) RouteQuery(ctx *astral.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	path, params := query.Parse(q.Query)

	op := set.Find(path)
	if op == nil {
		return query.RouteNotFound(set)
	}

	var query = newQuery(w, q)
	defer query.Reject()

	go func() {
		// ctx will end as soon as the query resolves, so we need a new one for the op
		ctx := astral.NewContext(nil).WithIdentity(ctx.Identity()).WithZone(astral.ZoneAll)

		// call the op
		err := op.Call(ctx, query, params)
		if err != nil && set.OnError != nil {
			set.OnError(err, q)
		}

		// reject the query in case the op did not respond to it, will do nothing if it did.
		_ = query.Reject()
	}()

	return query.resolve(ctx)
}
