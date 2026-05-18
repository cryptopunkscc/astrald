# Concurrency Patterns

Use these rules when choosing synchronization primitives or handling shared
mutable state.

## Mutex

- Name mutex fields `mu`; never embed them.
- Put `defer Unlock()` on the same line as `Lock()`.
- Use `sync.RWMutex` for read-heavy state.

```go
type Module struct {
    mu    sync.RWMutex
    items map[string]*Thing
}

func (m *Module) Add(key string, t *Thing) { m.mu.Lock(); defer m.mu.Unlock(); m.items[key] = t }
func (m *Module) Find(key string) *Thing   { m.mu.RLock(); defer m.mu.RUnlock(); return m.items[key] }
```

Source: `sig/map.go`, `core/conn.go`

## Atomic

Use atomics for lock-free flags, state machines, counters, and timestamps.

```go
func (c *conn) Close() error {
    if !c.closed.CompareAndSwap(false, true) {
        return nil
    }
    c.mu.Lock(); defer c.mu.Unlock()
    return nil
}
```

| Type | Use |
|---|---|
| `atomic.Bool` | closed/done/enabled flags |
| `atomic.Int32` | state machine |
| `atomic.Uint64` | byte counters, sequence numbers |
| `atomic.Int64` | Unix nanosecond timestamps |

Source: `core/conn.go`, `mod/nodes/src/session.go`, `mod/nat/src/hole.go`

## sync.Cond

- Use `sync.Cond` only when a goroutine must block on a computed condition.
- Call `.Wait()` inside a `for` loop.

```go
c.wcond.L.Lock()
defer c.wcond.L.Unlock()
for c.wsize == 0 {
    c.wcond.Wait()
}
```

Source: `mod/nodes/src/session.go`

## WaitGroup

- Call `Add(1)` before `go`.
- Make `defer wg.Done()` the first statement in the goroutine.
- Use a local variable, not a struct field.

```go
var wg sync.WaitGroup
var errCh = make(chan error, 32)

wg.Add(1)
go func() {
    defer wg.Done()
    if err := doWork(ctx); err != nil {
        errCh <- err
    }
}()

wg.Wait()
```

Source: `core/run.go`

## Done And Error Channels

- Signal completion by closing a channel, not by sending.

```go
done := make(chan struct{})
go func() {
    defer close(done)
    doWork()
}()
<-done
```

- Buffer error channels with capacity at least equal to the number of senders.
- Use `sig` helpers for context-aware channel operations.

```go
err := sig.RecvErr(ctx, errc)
v, err := sig.Recv[T](ctx, ch)
err = sig.Send(ctx, ch, value)
```

Source: `sig/signal.go`, `sig/chan.go`, `core/run.go`

## sig Collections

Prefer `sig.Map`, `sig.Set`, and `sig.Queue` over raw mutex plus map/slice for
shared mutable state.

```go
m := sig.Map[string, *Thing]{}
m.Set("key", thing)
v, ok := m.Get("key")
all := m.Clone()

s := sig.Set[*Thing]{}
s.Add(thing)
s.Remove(thing)

q := sig.Queue[*Event]{}
q.Push(event)
event := q.Next()
```

Source: `sig/`
