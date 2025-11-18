package nat

type PairState int32

const (
	StateIdle      PairState = iota // normal keepalive
	StateInLocking                  // lock requested, waiting for drain
	StateLocked                     // socket silent, no traffic
	StateExpired                    // mapping corrupted / lack of reachability
)
