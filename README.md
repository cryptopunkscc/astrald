astrald
=======

### Status

This is a proof-of-concept implementation of astral - a general-purpose peer-to-peer network. Like most PoCs,
this code is hightly experimental and will break and crash and might not work for you at all. At this stage
I publish this for developers to get early feedback on the ideas expressed by the code.

### Getting the node running

Get the code:
```
$ git clone https://github.com/cryptopunkscc/astrald.git
Cloning into 'astrald'...
...
$ cd astrald
```
Install all commands:
```
$ go install ./cmd/...
```
Start the node:
```
$ astrald
```

### Build an app

Minimal server example:
```go
package main

import "github.com/cryptopunkscc/astrald/lib/astral"

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

Minimal client example:
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

### Using tor

You'll need Tor with control port and cookie authentication. If you haven't already, add these lines to /etc/tor/torrc:

```
ControlPort 9051
CookieAuthentication 1
```

And restart tor and then astrald.

### Contact

You can reach me via email or XMPP: arashi@cryptopunks.cc