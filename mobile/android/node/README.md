# Astral Mobile Node

Astral runner library for mobile.

# API

* Start - Run node. Only one instance can be run at time. The function is blocking.
* Stop - Stop already running node.
* Identity - Returns identity of running node.

## Build Android ARR

Example command for generating aar and sources lib:

```shell
gomobile bind -v -o ./astral.aar -target=android .
```

### NOTE

The build may fail if go-mobile is not included in go.mod. To add go-mobile to dependencies run:

```shell
go get golang.org/x/mobile/bind
```

## References

* https://github.com/golang/go/wiki/Mobile#building-and-deploying-to-android-1
* https://pkg.go.dev/golang.org/x/mobile/cmd/gobind#hdr-Binding_Go
