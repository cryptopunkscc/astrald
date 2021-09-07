# Astral Android

Basic compatibility layer for Android

## Build

Run the following command from the project root to generate aar and sources lib:

```shell
gomobile bind -v -o $OUTPUT_DIR -target=android ./bind/android/
```

## References

* https://github.com/golang/go/wiki/Mobile#building-and-deploying-to-android-1
* https://pkg.go.dev/golang.org/x/mobile/cmd/gobind#hdr-Binding_Go
