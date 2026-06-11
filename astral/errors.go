package astral

import (
	"errors"
)

// simple errors

// ErrTimeout - query timed out
var ErrTimeout = errors.New("query timeout")

// ErrZoneExcluded - operation requires zones excluded from the scope
var ErrZoneExcluded = errors.New("zone excluded")

// ErrTargetNotAllowed - target was blocked by a policy or a filter
var ErrTargetNotAllowed = errors.New("target not allowed")

// blueprint registration errors

// ErrAlreadyRegistered - a Blueprint with the same Type is already registered (compile-time or runtime)
var ErrAlreadyRegistered = errors.New("blueprint already registered")

// ErrBlueprintNotFound - no Blueprint or compile-time prototype is registered for the named Object Type
var ErrBlueprintNotFound = errors.New("blueprint not found")

// ErrPrimitiveType - the named Type is a wire primitive; primitives have no Blueprint
var ErrPrimitiveType = errors.New("primitive type has no blueprint")

// ErrStreamCorrupted - an unknown type tag was consumed mid-stream and the remaining bytes
// cannot be safely interpreted. Wraps ErrBlueprintNotFound at sites that consumed bytes they
// cannot skip (Decode, interfaceValue.ReadFrom). Channel receivers without per-object framing
// latch on it and refuse subsequent reads.
var ErrStreamCorrupted = errors.New("stream corrupted")

// ErrBlueprintInvalid - the Blueprint failed structural or allowlist validation
var ErrBlueprintInvalid = errors.New("invalid blueprint")

// ErrFieldTypeMismatch - RuntimeObject.Set received a value that does not match the field's Spec
var ErrFieldTypeMismatch = errors.New("field type mismatch")

// ErrDepthExceeded - RuntimeObject encode/decode exceeded MaxBlueprintDepth nested frames
var ErrDepthExceeded = errors.New("blueprint depth exceeded")

// ErrBlueprintCycle - reachable when the topo-sort over registered Blueprints can't make
// progress (a reference cycle). Production usually can't construct one — RegisterBlueprint's
// validateReferences rejects forward refs — but a parent-after-child shadow can introduce
// one. Callers receive both the cycle-bearing list (with the stuck nodes appended in alpha
// order) and the error, so sync can still replay best-effort while logging the divergence.
var ErrBlueprintCycle = errors.New("blueprint reference cycle")
