# astrald

A client library for `astrald`.

## Quick start

### Sending a query

```go
package main

import (
	"fmt"

	"github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/lib/query"
)

func main() {
	ch, err := astrald.QueryChannel("target", "method", query.Args{"arg": "val"})
	if err != nil {
		panic(err)
	}

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

import (
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/astrald"
)

func main() {
	l, err := astrald.Listen()
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

import (
	"fmt"
	"os"

	"github.com/cryptopunkscc/astrald/lib/astrald"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: astral-resolve <alias>")
		return
	}

	identity, err := astrald.Dir().ResolveIdentity(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return
	}

	fmt.Println(identity)
}
```