### Warpdrive CLI thorug ANC

Combined with [ANC](../anc/README.md) (astral netcat) can be considered as
referential UI client for development and testing purpose. Under the hood, it translates text commands directly into
warpdrive UI API calls and returns received results as formatted console output. Depending on needs is recommended to
use more than one client instance connected at the same time, this can be useful for tracking subscription updates and
triggering updates at the same time.

### Running console client

ANC is a commandline astral client, that allows querying services registered under specified port, on local or remote
device:

```shell
$ anc query <identity> <port>
```

For example, to connect to warpdrive CLI on localnode, you can run ANC directly from package [amc](../anc):

```shell
$ go run ./cmd/anc query localnode wd
connected.
```