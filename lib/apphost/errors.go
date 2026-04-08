package apphost

import "errors"

// ErrNodeUnavailable is returned when the IPC connection to the node cannot be
// established. RetryRouter keys on this error to decide whether to retry —
// custom Router wrappers must return it (or wrap it) to participate in retry logic.
var ErrNodeUnavailable = errors.New("node unavailable")
