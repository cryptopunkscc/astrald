package nodes

import (
	"testing"
	"time"
)

// testCfg returns a config with small, round numbers for deterministic tests.
// LeakRate=1000 B/s, LevelRef=5000 B, so level hits LevelRef after 5 seconds of
// net fill. WLevel=1.0 means level-only score == 1.0 at LevelRef. WRTT=1.0
// means RTT-only score == 1.0 at RTTRef.
func testCfg() StreamPressureConfig {
	return StreamPressureConfig{
		LeakRate: 1000,
		Cap:      10_000,
		LevelRef: 5_000,
		RTTRef:   100 * time.Millisecond,
		RTTAlpha: 0.5,
		WLevel:   1.0,
		WRTT:     1.0,
		Enter:    1.0,
		Exit:     0.4,
	}
}

var epoch = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

// TestSustainedThroughputTriggersOnce: filling bucket to LevelRef fires once;
// additional bytes while HIGH do not fire again.
func TestSustainedThroughputTriggersOnce(t *testing.T) {
	var calls int
	pd := NewStreamPressureDetector(epoch, testCfg(), func() { calls++ })

	// t+1ms: decay=1 byte, then add 5001 → level=5000 → score=1.0 → HIGH
	pd.OnBytes(5001, epoch.Add(time.Millisecond))
	if calls != 1 {
		t.Fatalf("expected 1 call after entering HIGH, got %d", calls)
	}

	// still HIGH, additional bytes must not retrigger
	pd.OnBytes(1000, epoch.Add(2*time.Millisecond))
	if calls != 1 {
		t.Fatalf("expected still 1 call while HIGH, got %d", calls)
	}
}

// TestBurstThenIdleDecays: after entering HIGH, a long idle period drains the
// bucket below Exit, resetting the gate. A subsequent burst retriggers.
func TestBurstThenIdleDecays(t *testing.T) {
	var calls int
	pd := NewStreamPressureDetector(epoch, testCfg(), func() { calls++ })

	// enter HIGH at t+1ms
	pd.OnBytes(5001, epoch.Add(time.Millisecond))
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}

	// idle for 10 s: decay = 1000*10 = 10000 B → level floors to 0 → score=0 → exits HIGH
	pd.OnBytes(0, epoch.Add(10*time.Second))

	// re-enter HIGH: must fire again
	pd.OnBytes(5001, epoch.Add(10*time.Second+time.Millisecond))
	if calls != 2 {
		t.Fatalf("expected 2 calls after re-entry, got %d", calls)
	}
}

// TestHighRTTAloneTriggers: with zero throughput, a single RTT sample at RTTRef
// drives score to 1.0 and fires the callback.
func TestHighRTTAloneTriggers(t *testing.T) {
	var calls int
	pd := NewStreamPressureDetector(epoch, testCfg(), func() { calls++ })

	// first sample seeds rttEma directly; rttNorm = 100ms/100ms = 1.0 → score=1.0
	pd.OnRTT(100*time.Millisecond, epoch.Add(time.Second))
	if calls != 1 {
		t.Fatalf("expected 1 call on high RTT, got %d", calls)
	}
}

// TestOscillationNearThresholdsNoRetrigger: score bouncing around Enter/Exit
// while remaining in the HIGH zone must not produce repeated callbacks.
func TestOscillationNearThresholdsNoRetrigger(t *testing.T) {
	var calls int
	pd := NewStreamPressureDetector(epoch, testCfg(), func() { calls++ })

	now := epoch.Add(time.Millisecond)

	// enter HIGH
	pd.OnBytes(5001, now)
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}

	// oscillate: add bytes that push score high, then let decay bring it back—
	// but never cross Exit (0.4 → level < 2000). Score stays above Exit.
	for i := 0; i < 10; i++ {
		now = now.Add(500 * time.Millisecond) // decay 500 B; level stays well above 2000
		pd.OnBytes(500, now)                  // refill; score remains in HIGH zone
	}

	if calls != 1 {
		t.Fatalf("expected still 1 call during oscillation, got %d", calls)
	}
}

// TestLongIdleGapDecaysCorrectly: a single call after a very long gap must apply
// full decay in one step, not require intermediate calls.
func TestLongIdleGapDecaysCorrectly(t *testing.T) {
	var calls int
	pd := NewStreamPressureDetector(epoch, testCfg(), func() { calls++ })

	// enter HIGH
	pd.OnBytes(5001, epoch.Add(time.Millisecond))
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}

	// one call after 1 hour of idle: level must floor to 0, gate resets
	pd.OnBytes(0, epoch.Add(time.Hour))

	// fresh burst must retrigger exactly once
	pd.OnBytes(5001, epoch.Add(time.Hour+time.Millisecond))
	if calls != 2 {
		t.Fatalf("expected 2 calls after long idle + re-entry, got %d", calls)
	}
}

// TestNilSafe: Link with nil detector must not panic.
func TestNilSafe(t *testing.T) {
	s := &Link{}
	if s.pressure != nil {
		s.pressure.OnBytes(1024, time.Now())
		s.pressure.OnRTT(100*time.Millisecond, time.Now())
	}
}
