package astral

const (
	CodeSuccess = iota
	CodeRejected
	CodeInvalidQuery
	CodeCanceled
	CodeInternalError
)

const DefaultRejectCode = 1
