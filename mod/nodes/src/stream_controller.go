package nodes

import (
	"github.com/cryptopunkscc/astrald/sig"
)

// StreamController owns the execution of stream policy actions and protection state.
// It implements StreamControl and is a separate responsibility from StreamManager.
type StreamController struct {
	protected sig.Set[int] // set of protected stream IDs (per-run)
}

func NewStreamController() *StreamController {
	return &StreamController{}
}

func (sc *StreamController) ProtectStream(s *Stream) error {
	_ = sc.protected.Add(s.id)
	return nil
}

func (sc *StreamController) CloseStream(s *Stream) error {
	if sc.IsProtected(s) {
		return nil
	}
	return s.CloseWithError(nil)
}

func (sc *StreamController) IsProtected(s *Stream) bool {
	return sc.protected.Contains(s.id)
}

// ClearProtection removes protection from a stream (useful for periodic cleanup)
func (sc *StreamController) ClearProtection(s *Stream) {
	_ = sc.protected.Remove(s.id)
}

// ClearAllProtections clears all stream protections (useful for periodic re-evaluation)
func (sc *StreamController) ClearAllProtections() {
	_ = sc.protected.Clear()
}
