# Astral Android

Basic compatibility layer for Android

## Build

Following command will generate aar and sources lib:

```shell
gomobile bind -v -o $OUTPUT_DIR -target=android ./mobile/ ./java/
```

## References

* https://github.com/golang/go/wiki/Mobile#building-and-deploying-to-android-1
* https://pkg.go.dev/golang.org/x/mobile/cmd/gobind#hdr-Binding_Go
