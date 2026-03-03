package nodes

import "time"

type PressureDetector interface {
	OnBytes(n int, now time.Time)
	OnRTT(rtt time.Duration, now time.Time)
}

type PressureConfig struct {
	LeakRate float64 // bytes/sec decay
	Cap      float64 // max backlog
	LevelRef float64 // normalization reference for level

	RTTRef   time.Duration // normalization reference for RTT
	RTTAlpha float64       // EMA smoothing factor (0..1)

	WLevel float64 // weight of level component
	WRTT   float64 // weight of RTT component

	Enter float64 // score threshold to enter HIGH
	Exit  float64 // score threshold to exit HIGH (must be < Enter)
}

var DefaultPressureConfig = PressureConfig{
	LeakRate: 50 * 1024,
	Cap:      500 * 1024,
	LevelRef: 200 * 1024,

	RTTRef:   200 * time.Millisecond,
	RTTAlpha: 0.25,

	WLevel: 0.7,
	WRTT:   0.5,

	Enter: 1.0,
	Exit:  0.4,
}

func NewPressureDetector(now time.Time, cfg PressureConfig, onHigh func()) PressureDetector {
	return &pressureDetector{
		lastUpdate: now,
		cfg:        cfg,
		onHigh:     onHigh,
	}
}

type pressureDetector struct {
	cfg        PressureConfig
	onHigh     func()
	level      float64
	rttEma     float64
	lastUpdate time.Time
	high       bool
}

func (p *pressureDetector) decay(now time.Time) {
	dt := now.Sub(p.lastUpdate).Seconds()
	p.lastUpdate = now
	p.level -= p.cfg.LeakRate * dt
	if p.level < 0 {
		p.level = 0
	}
}

func (p *pressureDetector) score() float64 {
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

func (p *pressureDetector) gate(s float64) {
	if !p.high && s >= p.cfg.Enter {
		p.high = true
		p.onHigh()
		return
	}
	if p.high && s <= p.cfg.Exit {
		p.high = false
	}
}

func (p *pressureDetector) OnBytes(n int, now time.Time) {
	p.decay(now)
	p.level += float64(n)
	if p.level > p.cfg.Cap {
		p.level = p.cfg.Cap
	}
	p.gate(p.score())
}

func (p *pressureDetector) OnRTT(rtt time.Duration, now time.Time) {
	p.decay(now)
	if p.rttEma == 0 {
		p.rttEma = float64(rtt)
	} else {
		p.rttEma = p.cfg.RTTAlpha*float64(rtt) + (1-p.cfg.RTTAlpha)*p.rttEma
	}
	p.gate(p.score())
}
