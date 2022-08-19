# Astral mobile node

The library module, wraps astral node and provides as android library.

## Android astral API

The module provides minimal set of methods required for managing astral node on android.

* Start astral node. Only one instance can be run at time. The function is blocking.
* Stop already running node.
* Obtain identity of running node.

## Android native API

In some cases astral node executed on android device can require access to native android API. According
to [gomobile](https://pkg.go.dev/golang.org/x/mobile/cmd/gobind#hdr-Passing_target_language_objects_to_Go)
, this can be provided as interface defined in golang and exposed as java interface through astral node arr library to
be implemented on android side.

### Custom interface

For exposing some android API using custom interface is required to:

1. Declare golang interface inside [node](../node)` directory.
2. Expose as argument of [Start](main.go) function.
3. Implement exposed java interface in [node wrapper module](src/main/java/cc/cryptopunks/astral/wrapper).
4. Pass interface implementation to [Node.start](src/main/java/cc/cryptopunks/astral/wrapper/Astral.kt) method call.

## Astral node ARR

This library module requires astral node arr as dependency.

If you want generate astral arr manually, execute following command:

```shell
gomobile bind -v -o ../build/astral.aar -target=android .
```

or use predefined [script](../buildGo.sh) from parent directory.

### NOTE

The build may fail if go-mobile is not included in go.mod. To add go-mobile to dependencies run:

```shell
go get golang.org/x/mobile/bind
```

## References

* https://github.com/golang/go/wiki/Mobile#building-and-deploying-to-android-1
* https://pkg.go.dev/golang.org/x/mobile/cmd/gobind#hdr-Binding_Go
