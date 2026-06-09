package astral

import (
	"fmt"
	"io"
)

// objectReader / objectWriter wrap io.Reader / io.Writer to carry per-call state across
// nested RuntimeObject.ReadFrom / WriteTo calls. RuntimeObject wraps on entry; nested frames
// detect the wrapper and reuse it. Primitives that only need io.Reader / io.Writer don't
// notice the wrapping. The reader additionally carries the per-call Blueprints registry used
// to resolve type names referenced from a Spec; the writer needs no registry because WriteTo
// holds the concrete Object whose schema is already bound.

type objectReader struct {
	io.Reader
	depth int
	// bps is the registry that nested field reads consult to resolve type names referenced
	// from a Spec (PrimitiveSpec, RefSpec, PtrSpec). Nil → defaultBlueprints, matching the
	// pre-WithBlueprints behaviour. Decode threads cfg.Blueprints in here so a per-call
	// registry (set via WithBlueprints) flows down into every nested ReadFrom frame.
	bps *Blueprints
}

type objectWriter struct {
	io.Writer
	depth int
}

// resolve returns the registry the reader should consult for nested type lookups.
// Falls back to defaultBlueprints so paths that hand-wrap a reader without setting bps
// keep their historical behaviour.
func (or *objectReader) resolve() *Blueprints {
	if or.bps != nil {
		return or.bps
	}
	return defaultBlueprints
}

// enter records entry into a nested frame and fails with ErrDepthExceeded once the cap is
// crossed. Pair with a deferred exit() so the counter unwinds on every return path, including
// the error path.
func (ow *objectWriter) enter(typ fmt.Stringer) error {
	ow.depth++
	if ow.depth > MaxBlueprintDepth {
		return fmt.Errorf("%w: %s", ErrDepthExceeded, typ)
	}
	return nil
}

func (ow *objectWriter) exit() { ow.depth-- }

func (or *objectReader) enter(typ fmt.Stringer) error {
	or.depth++
	if or.depth > MaxBlueprintDepth {
		return fmt.Errorf("%w: %s", ErrDepthExceeded, typ)
	}
	return nil
}

func (or *objectReader) exit() { or.depth-- }
