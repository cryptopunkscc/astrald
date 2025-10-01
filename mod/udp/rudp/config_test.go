package rudp

import (
	"testing"
	"time"
)

// TestNormalizeAppliesDefaults verifies that a zero-value Config gets all default values.
func TestNormalizeAppliesDefaults(t *testing.T) {
	var c Config
	c.Normalize()

	if c.MaxSegmentSize != DefaultMSS {
		f(t, "MaxSegmentSize", c.MaxSegmentSize, DefaultMSS)
	}
	if c.MaxWindowBytes != DefaultWindowBytes {
		f(t, "MaxWindowBytes", c.MaxWindowBytes, DefaultWindowBytes)
	}
	if c.MaxWindowPackets != DefaultWndPkts {
		f(t, "MaxWindowPackets", c.MaxWindowPackets, DefaultWndPkts)
	}
	if c.RetransmissionInterval != DefaultRTO {
		f(t, "RetransmissionInterval", c.RetransmissionInterval, DefaultRTO)
	}
	if c.MaxRetransmissionInterval != DefaultRTOMax {
		f(t, "MaxRetransmissionInterval", c.MaxRetransmissionInterval, DefaultRTOMax)
	}
	if c.RetransmissionLimit != DefaultRetries {
		f(t, "RetransmissionLimit", c.RetransmissionLimit, DefaultRetries)
	}
	if c.AckDelay != DefaultAckDelay {
		f(t, "AckDelay", c.AckDelay, DefaultAckDelay)
	}
	if c.RecvBufBytes != DefaultRecvBufBytes {
		f(t, "RecvBufBytes", c.RecvBufBytes, DefaultRecvBufBytes)
	}
	if c.SendBufBytes != DefaultSendBufBytes {
		f(t, "SendBufBytes", c.SendBufBytes, DefaultSendBufBytes)
	}
}

// TestNormalizePreservesNonZero verifies that non-zero fields are not overwritten.
func TestNormalizePreservesNonZero(t *testing.T) {
	orig := Config{
		MaxSegmentSize:            999,
		MaxWindowBytes:            123456,
		MaxWindowPackets:          77,
		RetransmissionInterval:    321 * time.Millisecond,
		MaxRetransmissionInterval: 987 * time.Millisecond,
		RetransmissionLimit:       42,
		AckDelay:                  11 * time.Millisecond,
		RecvBufBytes:              222222,
		SendBufBytes:              333333,
	}
	c := orig
	c.Normalize()

	if c != orig {
		// Compare field-by-field for clearer diagnostics.
		if c.MaxSegmentSize != orig.MaxSegmentSize {
			g(t, "MaxSegmentSize", c.MaxSegmentSize, orig.MaxSegmentSize)
		}
		if c.MaxWindowBytes != orig.MaxWindowBytes {
			g(t, "MaxWindowBytes", c.MaxWindowBytes, orig.MaxWindowBytes)
		}
		if c.MaxWindowPackets != orig.MaxWindowPackets {
			g(t, "MaxWindowPackets", c.MaxWindowPackets, orig.MaxWindowPackets)
		}
		if c.RetransmissionInterval != orig.RetransmissionInterval {
			g(t, "RetransmissionInterval", c.RetransmissionInterval, orig.RetransmissionInterval)
		}
		if c.MaxRetransmissionInterval != orig.MaxRetransmissionInterval {
			g(t, "MaxRetransmissionInterval", c.MaxRetransmissionInterval, orig.MaxRetransmissionInterval)
		}
		if c.RetransmissionLimit != orig.RetransmissionLimit {
			g(t, "RetransmissionLimit", c.RetransmissionLimit, orig.RetransmissionLimit)
		}
		if c.AckDelay != orig.AckDelay {
			g(t, "AckDelay", c.AckDelay, orig.AckDelay)
		}
		if c.RecvBufBytes != orig.RecvBufBytes {
			g(t, "RecvBufBytes", c.RecvBufBytes, orig.RecvBufBytes)
		}
		if c.SendBufBytes != orig.SendBufBytes {
			g(t, "SendBufBytes", c.SendBufBytes, orig.SendBufBytes)
		}
		// Fail after reporting discrepancies.
		if t.Failed() {
			return
		}
	}
}

// TestNormalizePartial ensures only zero fields get populated.
func TestNormalizePartial(t *testing.T) {
	c := Config{
		MaxSegmentSize: 500, // keep
		// others zero
		AckDelay: 5 * time.Millisecond, // keep
	}
	c.Normalize()

	if c.MaxSegmentSize != 500 {
		g(t, "MaxSegmentSize", c.MaxSegmentSize, 500)
	}
	if c.AckDelay != 5*time.Millisecond {
		g(t, "AckDelay", c.AckDelay, 5*time.Millisecond)
	}

	if c.MaxWindowBytes != DefaultWindowBytes {
		f(t, "MaxWindowBytes", c.MaxWindowBytes, DefaultWindowBytes)
	}
	if c.MaxWindowPackets != DefaultWndPkts {
		f(t, "MaxWindowPackets", c.MaxWindowPackets, DefaultWndPkts)
	}
	if c.RetransmissionInterval != DefaultRTO {
		f(t, "RetransmissionInterval", c.RetransmissionInterval, DefaultRTO)
	}
	if c.MaxRetransmissionInterval != DefaultRTOMax {
		f(t, "MaxRetransmissionInterval", c.MaxRetransmissionInterval, DefaultRTOMax)
	}
	if c.RetransmissionLimit != DefaultRetries {
		f(t, "RetransmissionLimit", c.RetransmissionLimit, DefaultRetries)
	}
	if c.RecvBufBytes != DefaultRecvBufBytes {
		f(t, "RecvBufBytes", c.RecvBufBytes, DefaultRecvBufBytes)
	}
	if c.SendBufBytes != DefaultSendBufBytes {
		f(t, "SendBufBytes", c.SendBufBytes, DefaultSendBufBytes)
	}
}

// TestNormalizeIdempotent ensures calling Normalize twice doesn't change values after first call.
func TestNormalizeIdempotent(t *testing.T) {
	var c Config
	c.Normalize()
	first := c
	c.Normalize()
	if c != first {
		g(t, "ConfigAfterSecondNormalize", c, first)
	}
}

// TestNormalizeNegativeValues ensures negative values are preserved (no implicit clamping yet).
func TestNormalizeNegativeValues(t *testing.T) {
	c := Config{
		MaxSegmentSize:      -1,
		MaxWindowBytes:      -2,
		MaxWindowPackets:    -3,
		RetransmissionLimit: -4,
		RecvBufBytes:        -5,
		SendBufBytes:        -6,
	}
	// Durations negative as well
	c.RetransmissionInterval = -10 * time.Millisecond
	c.MaxRetransmissionInterval = -20 * time.Millisecond
	c.AckDelay = -30 * time.Millisecond

	c.Normalize()

	if c.MaxSegmentSize != -1 {
		g(t, "MaxSegmentSize", c.MaxSegmentSize, -1)
	}
	if c.MaxWindowBytes != -2 {
		g(t, "MaxWindowBytes", c.MaxWindowBytes, -2)
	}
	if c.MaxWindowPackets != -3 {
		g(t, "MaxWindowPackets", c.MaxWindowPackets, -3)
	}
	if c.RetransmissionLimit != -4 {
		g(t, "RetransmissionLimit", c.RetransmissionLimit, -4)
	}
	if c.RecvBufBytes != -5 {
		g(t, "RecvBufBytes", c.RecvBufBytes, -5)
	}
	if c.SendBufBytes != -6 {
		g(t, "SendBufBytes", c.SendBufBytes, -6)
	}
	if c.RetransmissionInterval != -10*time.Millisecond {
		g(t, "RetransmissionInterval", c.RetransmissionInterval, -10*time.Millisecond)
	}
	if c.MaxRetransmissionInterval != -20*time.Millisecond {
		g(t, "MaxRetransmissionInterval", c.MaxRetransmissionInterval, -20*time.Millisecond)
	}
	if c.AckDelay != -30*time.Millisecond {
		g(t, "AckDelay", c.AckDelay, -30*time.Millisecond)
	}
}

// Helper failure formatters for brevity.
func f[T comparable](t *testing.T, field string, got, want T) {
	if got != want {
		// Using t.Fatalf to stop early in default-path tests where cascading errors add little value.
		// Use %v for generic print.
		//nolint:forbidigo // simple test diagnostic
		t.Fatalf("%s mismatch: got=%v want=%v", field, got, want)
	}
}

func g[T comparable](t *testing.T, field string, got, want T) {
	if got != want {
		//nolint:forbidigo
		t.Errorf("%s mismatch: got=%v want=%v", field, got, want)
	}
}
