## astral

Crude go library to use astrald via its apphost module.

### Quick start

### Server
```go
package main

import (
	"github.com/cryptopunkscc/astrald/lib/astral"
)

func main() {
	l, err := astral.Listen("myname.myapp")
	if err != nil {
		panic(err)
	}
	for {
		// l is a net.Listener
		conn, err := l.Accept()
		if err != nil {
			break
		}
		conn.Write([]byte("hello friend!\n"))
		conn.Close()
	}
}
```

### Client
```go
package main

import (
	"github.com/cryptopunkscc/astrald/lib/astral"
	"io"
	"os"
)

func main() {
	conn, err := astral.DialName("localnode", "myname.myapp")
	if err != nil {
		panic(err)
	}

	io.Copy(os.Stdout, conn)
}
```