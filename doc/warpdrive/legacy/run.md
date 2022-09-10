# Desktop

Manual how to run on warpdrive on desktop for testing and development purpose.

### Platforms

Operating system:

* Linux
* Mac? - not tested
* Windows? - not tested

### Requirements

Software:

* Golang
* Git

### Steps

For early development purpose the services and clients apps have to be run as separated process.

* Astral node - provides features like identity, connectivity, encryption and address book.
* Warpdrive service - handles communication with other warpdrive services on remote devices.
* UI application - the user interface for communication with warpdrive service on local device.

To run each required part, follow the steps:

#### 1. Ensure you are inside `warpdrive` branch in your local repository root.

If not, clone [repository](https://github.com/cryptopunkscc/astrald/tree/warpdrive) if needed and change directory in
terminal.

```shell
git clone git@github.com:cryptopunkscc/astrald.git
cd ./astrald
git checkout --track origin/warpdrive
```

#### 2. Ensure the astral node is running.

* If not, run `astrald` [./cmd/astrald/main.go](../../../cmd/astrald/main.go) executing:

```shell
go run ./cmd/astrald
```

#### 3. Execute the warpdrive service.

* From source [./app/warpdrive/cmd/main.go](../cmd/main.go)

```shell
go run ./app/warpdrive/cmd
```

#### 4. Both services now should be running.

Ensure there are no errors in console logs.

#### 5. Interacting with service

1. Standalone client - [CLI.md](../../../cmd/warpdrive/README.md)
2. ANC & `wd` service - [anc_cli.md](anc_cli.md)
