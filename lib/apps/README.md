# lib/apps

`lib/apps` registers an external app handler with the local node and routes
inbound apphost queries through an `astral.Router`.

`Serve(ctx, router, opts...)` creates a local IPC handler, registers it through
the default `AppRegistrar`, gates inbound queries until the registrar is ready,
then blocks routing queries until `ctx` is cancelled.

`ServeWith(ctx, router, registrar, opts...)` uses an explicit registrar. If the
registrar implements `lib/astrald.ReadyGate`, queries are gated until
`Ready()` closes. If serve options add registration hooks, the registrar must
implement `RegistrationHookRegistrar`.

## Basic app with routing.NewApp

`routing.NewApp` exposes exported methods with the op signature as routes.
Method names are converted to snake case, so `Hello` becomes `hello`.

```go
package main

import (
	"fmt"
	"os"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/apps"
	libastrald "github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/lib/routing"
)

type API struct{}

type helloArgs struct {
	Name string `query:"required"`
	Out  string `query:"optional"`
}

func (api *API) Hello(ctx *astral.Context, q *routing.IncomingQuery, args helloArgs) error {
	ch := q.Accept(channel.WithOutputFormat(args.Out))
	defer ch.Close()

	return ch.Send(astral.NewString16("hello, " + args.Name))
}

func main() {
	if err := apps.Serve(libastrald.NewContext(), routing.NewApp(&API{})); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

The app handles queries for `hello?name=alice`. `routing.NewApp` also exposes
`.spec`, which streams operation specs for the app.

To control registration explicitly, create a registrar and pass it to
`ServeWith`:

```go
ctx := libastrald.NewContext()
app := routing.NewApp(&API{})
reg := apps.NewAppRegistrar(ctx, apps.WithRetry())

if err := apps.ServeWith(ctx, app, reg); err != nil {
	panic(err)
}
```

`ServeWith` applies route mounts from serve options before registration, just
like `Serve`.

## Wiring configurations

Use these combinations for the common app shapes:

```go
// Plain app: routes only.
err := apps.Serve(ctx, routing.NewApp(api))

// App whose API implements objects.Finder: generate objects.find and register it.
err := apps.Serve(ctx, routing.NewApp(api), apps.WithObjectFinder(api))

// App whose API implements objects.Searcher: generate objects.search and register it.
err := apps.Serve(ctx, routing.NewApp(api), apps.WithObjectSearcher(api))

// App whose API implements objects.Describer: generate objects.describe and register it.
err := apps.Serve(ctx, routing.NewApp(api), apps.WithObjectDescriber(api))

// Combine several capabilities on the same app.
err := apps.Serve(ctx, routing.NewApp(api),
    apps.WithObjectFinder(api),
    apps.WithObjectSearcher(api),
    apps.WithObjectDescriber(api),
)

// App with its own objects.* implementation: mount the route yourself and
// register the app as an external provider via WithRegistrationHook.
app := routing.NewApp(api)
app.Add("objects", objectOps)
err := apps.Serve(ctx, app, apps.WithRegistrationHook(objectsclient.RegisterFinder))

// Explicit registrar: useful for retry policy, test doubles, or custom bind setup.
reg := apps.NewAppRegistrar(ctx, apps.WithRetry())
err := apps.ServeWith(ctx, routing.NewApp(api), reg)
```

## Generated objects.find route

`WithObjectFinder(finder)` mounts an `objects.find` route onto a
`routing.ScopedOpRouter` such as `routing.NewApp(...)`, and adds the replayable
registration hook that calls `objects.register_finder` on the local node.
The snippets below use `sig.Send` and `sig.RecvOk` so finder streams unblock
when the query context is cancelled.

```go
// App carries any regular routing ops. Left empty here for brevity.
type App struct{}

// MyFinder is a separate value that implements objects.Finder.
type MyFinder struct {
	ids map[string]struct{}
}

func (f *MyFinder) FindObject(ctx *astral.Context, id *astral.ObjectID) (<-chan *astral.Identity, error) {
	out := make(chan *astral.Identity)
	go func() {
		defer close(out)

		if id == nil {
			return
		}
		if _, ok := f.ids[id.String()]; !ok {
			return
		}
		if provider := ctx.Identity(); provider != nil && !provider.IsZero() {
			_ = sig.Send(ctx, out, provider)
		}
	}()
	return out, nil
}

func main() {
	app := &App{}
	finder := &MyFinder{ids: map[string]struct{}{
		"2e56f...": {},
	}}

	err := apps.Serve(
		libastrald.NewContext(),
		routing.NewApp(app),
		apps.WithObjectFinder(finder),
	)
	if err != nil {
		panic(err)
	}
}
```

`routing.NewApp(app)` and `apps.WithObjectFinder(finder)` take distinct
values: the first defines the routing surface (custom ops, `.spec`), the
second installs the finder capability. Nothing forces them to be the same
struct — they can be, but the two roles stay separable.

The generated route accepts `objects.find?id=<object-id>&out=<format>`, calls
each configured finder, streams non-zero provider identities, and finishes with
`astral.EOS`. Missing or zero `id` sends an Astral error followed by `EOS`.

## Generated objects.search route

`WithObjectSearcher(searcher)` mounts an `objects.search` route onto a
`routing.ScopedOpRouter`, and adds the replayable registration hook that calls
`objects.register_searcher` on the local node. The generated route fans the
query out across every configured searcher and forwards each result. Dedup
is left to the node-side `objects.search` op, which already dedupes across
all searchers.

```go
type App struct{}

// MySearcher is a separate value that implements objects.Searcher.
type MySearcher struct {
    index map[string]*astral.ObjectID
}

func (s *MySearcher) SearchObject(ctx *astral.Context, query objects.SearchQuery) (<-chan *objects.SearchResult, error) {
    out := make(chan *objects.SearchResult)
    go func() {
        defer close(out)

        needle := strings.ToLower(string(query.Query))
        for name, id := range s.index {
            if needle != "" && !strings.Contains(strings.ToLower(name), needle) {
                continue
            }
            _ = sig.Send(ctx, out, &objects.SearchResult{ObjectID: id})
        }
    }()
    return out, nil
}

func main() {
    app := &App{}
    searcher := &MySearcher{index: map[string]*astral.ObjectID{
        "alpha": &astral.ObjectID{ /* ... */ },
        "beta":  &astral.ObjectID{ /* ... */ },
    }}

    err := apps.Serve(
        libastrald.NewContext(),
        routing.NewApp(app),
        apps.WithObjectSearcher(searcher),
    )
    if err != nil {
        panic(err)
    }
}
```

The generated route accepts `objects.search?q=<query>&out=<format>`, parses
`q` via `objects.SearchQuery.UnmarshalText`, fans across every configured
searcher, drops `nil`/zero `ObjectID`s, and finishes with `astral.EOS`. An
empty `q` is a valid "match everything" query — passed through unchanged.
The node-side op dedupes across all searchers, so this route forwards
duplicates as-is.

## Generated objects.describe route

`WithObjectDescriber(describer)` mounts an `objects.describe` route onto a
`routing.ScopedOpRouter`, and adds the replayable registration hook that calls
`objects.register_describer` on the local node. The generated route fans the
request across every configured describer and forwards each non-empty
descriptor.

```go
type App struct{}

// MyDescriber is a separate value that implements objects.Describer.
type MyDescriber struct {
    titles map[string]string
}

func (d *MyDescriber) DescribeObject(ctx *astral.Context, id *astral.ObjectID) (<-chan *objects.Descriptor, error) {
    out := make(chan *objects.Descriptor)
    go func() {
        defer close(out)

        if id == nil || id.IsZero() {
            return
        }
        title, ok := d.titles[id.String()]
        if !ok {
            return
        }
        _ = sig.Send(ctx, out, &objects.Descriptor{
            ObjectID: id,
            Data:     astral.NewString16(title),
        })
    }()
    return out, nil
}

func main() {
    app := &App{}
    describer := &MyDescriber{titles: map[string]string{
        "2e56f...": "example object",
    }}

    err := apps.Serve(
        libastrald.NewContext(),
        routing.NewApp(app),
        apps.WithObjectDescriber(describer),
    )
    if err != nil {
        panic(err)
    }
}
```

The generated route accepts `objects.describe?id=<object-id>&out=<format>`,
fans across every configured describer, drops descriptors with `nil` `Data`,
and finishes with `astral.EOS`. Missing or zero `id` sends an Astral error
followed by `EOS`. `SourceID` is left to the describer; the node-side
`ExternalDescriber` stamps it from the registration identity later.

## Custom objects.* route plus registration

When the app exposes its own `objects.find`, `objects.search`, or
`objects.describe` route, register it as an external provider by passing the
matching `objectsclient.Register*` to `WithRegistrationHook`. The hook calls
the corresponding `objects.register_*` op on the local node after the app
handler is attached, and `AppRegistrar` replays it on reconnect. The example
below shows the Finder shape; Searcher and Describer follow the same
pattern.

```go
type objectOps struct {
	finder *API
}

type findArgs struct {
	ID  *astral.ObjectID
	Out string `query:"optional"`
}

func (ops *objectOps) Find(ctx *astral.Context, q *routing.IncomingQuery, args findArgs) error {
	ctx, cancel := ctx.WithCancel()
	defer cancel()

	ch := q.Accept(channel.WithOutputFormat(args.Out))
	defer ch.Close()

	if args.ID == nil || args.ID.IsZero() {
		if err := ch.Send(astral.NewError("id is required")); err != nil {
			return err
		}
		return ch.Send(&astral.EOS{})
	}

	providers, err := ops.finder.FindObject(ctx, args.ID)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	for {
		provider, ok, err := sig.RecvOk(ctx, providers)
		if err != nil || !ok {
			break
		}
		if provider != nil && !provider.IsZero() {
			if err := ch.Send(provider); err != nil {
				return err
			}
		}
	}
	return ch.Send(&astral.EOS{})
}

func main() {
	finder := &API{ids: map[string]struct{}{"2e56f...": {}}}

	app := routing.NewApp(&API{})
	app.Add("objects", &objectOps{finder: finder})

	err := apps.Serve(
		libastrald.NewContext(),
		app,
		apps.WithRegistrationHook(objectsclient.RegisterFinder),
	)
	if err != nil {
		panic(err)
	}
}
```

`app.Add("objects", ops)` scopes `Find` under `objects`, so the method handles
`objects.find`. The registration hook makes the node add this app as an
external finder, causing the node's own `objects.find` implementation to query
the app later.

## Custom replay hook

Registration hooks run after the app handler is attached to the node. The
default `AppRegistrar` also runs them after reconnect replay, after all known
handlers have been re-attached.

Use hooks for plug-and-play registrations the node forgets when the bind
channel drops — external finders, describers, and searchers. Do not use
them for registrations the node persists in its own state (for example
indexers); those survive reconnects without replay.

```go
func main() {
	app := routing.NewApp(&API{})

	err := apps.Serve(
		libastrald.NewContext(),
		app,
		apps.WithRegistrationHook(objectsclient.RegisterDescriber),
		apps.WithRegistrationHook(objectsclient.RegisterSearcher),
	)
	if err != nil {
		panic(err)
	}
}
```

`WithRegistrationHooks(h1, h2, ...)` adds several hooks in one call. Nil
serve options or nil hooks return an error before serving.
