package event

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/sig"
	"reflect"
	"sync"
)

type Event interface{}

type Queue struct {
	mu     sync.Mutex
	queue  *sig.Queue
	parent *Queue
}

var ErrReturn = errors.New("fn must return exactly one error value")
var ErrArgument = errors.New("fn must take exactly one argument")
var ErrNotAFunc = errors.New("fn is not a function")

func (q *Queue) Emit(event Event) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.queue == nil {
		q.queue = &sig.Queue{}
	}

	q.queue = q.queue.Push(event)
	if q.parent != nil {
		q.parent.Emit(event)
	}
}

func (q *Queue) Subscribe(ctx context.Context) <-chan Event {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.queue == nil {
		q.queue = &sig.Queue{}
	}

	var ch = make(chan Event)

	go func() {
		defer close(ch)
		for e := range q.queue.Subscribe(ctx) {
			ch <- e
		}
	}()

	return ch
}

// HandleFunc takes a one argument, one result function. It will subscribe to the queue and invoke fn for every item
// in the queue that matches the argument type. The function has to return a single value of type error. If the error
// is not nil, HandleFunc will return the error and stop processing the queue. When the context is done, HandleFunc
// returns nil.
func (q *Queue) HandleFunc(ctx context.Context, fn interface{}) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		fnType = reflect.TypeOf(fn)
		fnVal  = reflect.ValueOf(fn)
	)

	if fnType.Kind() != reflect.Func {
		return ErrNotAFunc
	}
	if fnType.NumIn() != 1 {
		return ErrArgument
	}
	if fnType.NumOut() != 1 {
		return ErrReturn
	}

	var (
		errType = reflect.TypeOf((*error)(nil)).Elem()
		retType = fnType.Out(0)
		argType = fnType.In(0)
	)

	if retType != errType {
		return ErrReturn
	}

	for event := range q.Subscribe(ctx) {
		evType := reflect.TypeOf(event)
		if evType.ConvertibleTo(argType) {
			ret := fnVal.Call([]reflect.Value{reflect.ValueOf(event)})
			if !ret[0].IsNil() {
				return ret[0].Interface().(error)
			}
		}
	}

	return nil
}

func (q *Queue) SetParent(parent *Queue) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q == parent {
		return
	}
	q.parent = parent
}

func (q *Queue) Parent() *Queue {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.parent
}
