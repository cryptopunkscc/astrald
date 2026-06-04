package astral

import (
	"fmt"
	"io"
)

// depthReader / depthWriter carry a recursion depth across nested RuntimeObject.ReadFrom /
// WriteTo calls. RuntimeObject wraps on entry; nested frames detect the wrapper and reuse it.
// Primitives that only need io.Reader / io.Writer don't notice the wrapping.

type depthReader struct {
	io.Reader
	depth int
}

type depthWriter struct {
	io.Writer
	depth int
}

// enter records entry into a nested frame and fails with ErrDepthExceeded once the cap is
// crossed. Pair with a deferred exit() so the counter unwinds on every return path, including
// the error path.
func (dw *depthWriter) enter(typ fmt.Stringer) error {
	dw.depth++
	if dw.depth > MaxBlueprintDepth {
		return fmt.Errorf("%w: %s", ErrDepthExceeded, typ)
	}
	return nil
}

func (dw *depthWriter) exit() { dw.depth-- }

func (dr *depthReader) enter(typ fmt.Stringer) error {
	dr.depth++
	if dr.depth > MaxBlueprintDepth {
		return fmt.Errorf("%w: %s", ErrDepthExceeded, typ)
	}
	return nil
}

func (dr *depthReader) exit() { dr.depth-- }
