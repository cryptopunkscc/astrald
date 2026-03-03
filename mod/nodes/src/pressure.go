package nodes

import "time"

type StreamPressureDetector interface {
	OnBytes(n int, now time.Time)
	OnRTT(rtt time.Duration, now time.Time)
	State() StreamPressureState
}

// StreamPressureState is a read-only snapshot of the detector at a given moment.
type StreamPressureState struct {
	Level float64       // current bucket fill after decay (bytes)
	RTT   time.Duration // smoothed round-trip time (EMA)
	Score float64       // combined weighted score (0..~3+)
	High  bool          // true while score is above the Enter threshold
}

type StreamPressureConfig struct {
	// LeakRate is how fast the token bucket drains in bytes/sec.
	// Traffic below this rate does not accumulate pressure; only sustained
	// throughput above it will fill the bucket toward LevelRef.
	LeakRate float64

	// Cap is the maximum bucket level in bytes.
	// Prevents the level score from growing unboundedly during long bursts.
	Cap float64

	// LevelRef is the bucket level in bytes that normalises to a level score
	// of 1.0. Together with WLevel it controls how much throughput pressure
	// alone can contribute to the trigger: trigger requires
	// WLevel*(level/LevelRef) + WRTT*(rttEma/RTTRef) >= Enter.
	LevelRef float64

	// RTTRef is the RTT that normalises to an RTT score of 1.0.
	// Set to the expected round-trip time of the current transport; anything
	// significantly above this pushes the score toward the Enter threshold.
	RTTRef time.Duration

	// RTTAlpha is the EMA smoothing factor for RTT samples (0..1).
	// Higher values react faster to latency changes; lower values filter
	// short spikes. Typical range: 0.1 (slow/stable) to 0.5 (fast/reactive).
	RTTAlpha float64

	// WLevel is the weight applied to the normalised bucket level in the score.
	// Increase to make sustained throughput the dominant upgrade trigger.
	WLevel float64

	// WRTT is the weight applied to the normalised RTT EMA in the score.
	// Increase to make latency degradation the dominant upgrade trigger.
	WRTT float64

	// Enter is the combined score threshold at which the detector fires onHigh
	// and enters the HIGH state. Must be greater than Exit.
	Enter float64

	// Exit is the score threshold below which the detector leaves the HIGH
	// state, re-arming the callback for the next HIGH entry. Must be < Enter.
	// The gap between Exit and Enter defines the hysteresis band that prevents
	// repeated triggers when the score oscillates near the boundary.
	Exit float64
}

// DefaultStreamPressureConfig is calibrated for transports that are clearly not
// the best available option (e.g. Tor, KCP/UDP hole-punch). It triggers when
// the stream sustains notably more traffic than its baseline or when round-trip
// latency climbs well above the transport norm.
var DefaultStreamPressureConfig = StreamPressureConfig{
	LeakRate: 50 * 1024,  // drain 50 KB/s; traffic above this fills the bucket
	Cap:      500 * 1024, // ceiling at ~10 s of burst headroom above LeakRate
	LevelRef: 200 * 1024, // 200 KB in bucket → level score of 1.0

	RTTRef:   200 * time.Millisecond, // baseline RTT; above this raises the score
	RTTAlpha: 0.25,                   // slow smoothing; ignores brief RTT spikes

	WLevel: 0.7, // throughput alone can trigger at ~1.43× LevelRef
	WRTT:   0.5, // RTT alone can trigger at 2× RTTRef

	Enter: 1.0, // fire when combined score reaches 1.0
	Exit:  0.4, // re-arm when score falls back below 0.4
}

func NewStreamPressureDetector(now time.Time, cfg StreamPressureConfig, onHigh func()) StreamPressureDetector {
	return &streamPressureDetector{
		lastUpdate: now,
		cfg:        cfg,
		onHigh:     onHigh,
	}
}

type streamPressureDetector struct {
	cfg        StreamPressureConfig
	onHigh     func()
	level      float64
	rttEma     float64
	lastUpdate time.Time
	high       bool
}

func (p *streamPressureDetector) decay(now time.Time) {
	dt := now.Sub(p.lastUpdate).Seconds()
	p.lastUpdate = now
	p.level -= p.cfg.LeakRate * dt
	if p.level < 0 {
		p.level = 0
	}
}

func (p *streamPressureDetector) score() float64 {
	levelNorm := p.level / p.cfg.LevelRef
	if levelNorm > 3 {
		levelNorm = 3
	}
	rttNorm := p.rttEma / float64(p.cfg.RTTRef)
	if rttNorm > 3 {
		rttNorm = 3
	}
	return p.cfg.WLevel*levelNorm + p.cfg.WRTT*rttNorm
}

func (p *streamPressureDetector) gate(s float64) {
	if !p.high && s >= p.cfg.Enter {
		p.high = true
		p.onHigh()
		return
	}
	if p.high && s <= p.cfg.Exit {
		p.high = false
	}
}

func (p *streamPressureDetector) OnBytes(n int, now time.Time) {
	p.decay(now)
	p.level += float64(n)
	if p.level > p.cfg.Cap {
		p.level = p.cfg.Cap
	}
	p.gate(p.score())
}

func (p *streamPressureDetector) State() StreamPressureState {
	dt := time.Now().Sub(p.lastUpdate).Seconds()
	level := p.level - p.cfg.LeakRate*dt
	if level < 0 {
		level = 0
	}
	levelNorm := min(level/p.cfg.LevelRef, 3)
	rttNorm := min(p.rttEma/float64(p.cfg.RTTRef), 3)
	return StreamPressureState{
		Level: level,
		RTT:   time.Duration(p.rttEma),
		Score: p.cfg.WLevel*levelNorm + p.cfg.WRTT*rttNorm,
		High:  p.high,
	}
}

func (p *streamPressureDetector) OnRTT(rtt time.Duration, now time.Time) {
	p.decay(now)
	if p.rttEma == 0 {
		p.rttEma = float64(rtt)
	} else {
		p.rttEma = p.cfg.RTTAlpha*float64(rtt) + (1-p.cfg.RTTAlpha)*p.rttEma
	}
	p.gate(p.score())
}
