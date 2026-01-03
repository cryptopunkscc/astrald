# astrald

A client library for `astrald`.

## Quick start

### Sending a query

```go
package main

func main() {
	// create a new context
	ctx := astrald.NewContext()
	
	fmt.Printf("authenticated as %s on %s\n",
		astrald.GuestID(),
		astrald.HostID(),
	)
	
	// query the localnode
	conn, err := astrald.Query(ctx, "method", query.Args{"arg": "val"})
	
	// ... or use a helper method to get a channel 
	ch, err := astrald.QueryChannel(ctx, "method", query.Args{"arg": "val"})
	
	// ... or query some other node
	nodeID, err := astrald.Dir().ResolveIdentity(ctx, "nodealias")
	
	conn, err = astrald.WithTarget(nodeID).Query(ctx, "method", nil)
	
	o, err := ch.Read()
	switch o.(type) {
	case nil: // error
		return

	default:
		fmt.Println(o)
	}
}
```

### Listening for queries

```go
package main

func main() {
	ctx := astrald.NewContext()
	
	l, err := astrald.AppHost().RegisterHandler(ctx)
	if err != nil {
		panic(err)
	}

	for {
		query, err := l.Next()
		if err != nil {
			return
		}

		fmt.Printf("Query %v from %v\n", query.Query(), query.Caller())

		ch := query.AcceptChannel()
		ch.Write(&astral.Ack{})
		ch.Close()
	}
}

```

### Resolving aliases

```go
package main

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: astral-resolve <alias>")
		return
	}
	
	ctx := astrald.NewContext()

	identity, err := astrald.Dir().ResolveIdentity(ctx, os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return
	}

	fmt.Println(identity)
}
```
