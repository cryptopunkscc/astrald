//go:build gomobilebind

// This file exists solely to keep `golang.org/x/mobile/bind` in go.mod so
// that `gomobile bind ./mobile` can find its generator dependency. The build
// tag prevents it from being compiled into regular builds.
package mobile

import _ "golang.org/x/mobile/bind"
