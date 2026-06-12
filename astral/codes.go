package astral

const (
	CodeSuccess = iota
	CodeRejected
	CodeInvalidQuery
	CodeCanceled
	CodeInternalError
)

// DefaultRejectCode is the code returned when a query is rejected without an explicit
// reason; mirrors CodeRejected.
const DefaultRejectCode = 1
