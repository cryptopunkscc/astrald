package nodes

// PolicyAction implementations for stream management

// ProtectStreamAction protects a stream from eviction
type ProtectStreamAction struct {
	Stream  *Stream
	Reasons []string // one or more reasons (typically policy names)
}

func NewProtectStreamAction(s *Stream, reasons ...string) *ProtectStreamAction {
	return &ProtectStreamAction{Stream: s, Reasons: dedupeReasons(reasons)}
}

func (a *ProtectStreamAction) Execute(ctrl StreamControl) error {
	return ctrl.ProtectStream(a.Stream)
}

// CloseStreamAction closes a stream
type CloseStreamAction struct {
	Stream  *Stream
	Reasons []string // one or more reasons (typically policy names)
}

func NewCloseStreamAction(s *Stream, reasons ...string) *CloseStreamAction {
	return &CloseStreamAction{Stream: s, Reasons: dedupeReasons(reasons)}
}

func (a *CloseStreamAction) Execute(ctrl StreamControl) error {
	return ctrl.CloseStream(a.Stream)
}

// dedupeReasons removes duplicates while preserving order
func dedupeReasons(rs []string) []string {
	if len(rs) <= 1 {
		return rs
	}
	seen := make(map[string]struct{}, len(rs))
	out := make([]string, 0, len(rs))
	for _, r := range rs {
		if r == "" {
			continue
		}
		if _, ok := seen[r]; ok {
			continue
		}
		seen[r] = struct{}{}
		out = append(out, r)
	}
	return out
}
