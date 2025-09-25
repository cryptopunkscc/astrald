package udp

import (
	"time"
)

// RFC-backed constants for src UDP config
const (
	ListenPort = 1791
	// QUIC requires endpoints to handle 1200-byte UDP datagrams without fragmentation (RFC 9000 §14.1)

	DefaultMSS = 1200 - 13 // 1187: 1200 minus our header
	MinMSS     = 512       // RFC 8085: avoid fragmentation, safe for most links
	MaxMSS     = 1400      // Keeps under 1500B MTU with IP/UDP/tunnel headroom (RFC 8085, RFC 791, RFC 8200)

	// WindowBytes conservative buffer, RFC 8085

	DefaultWindowBytes = 16 * DefaultMSS
	MinWindowBytes     = MinMSS
	MaxWindowBytes     = 1 << 20 // 1 MiB

	// Retransmission timers: RFC 6298

	DefaultRTO    = 500 * time.Millisecond
	DefaultRTOMax = 4 * time.Second
	MinRTO        = 10 * time.Millisecond // LAN-friendly floor
	MaxRTOCeiling = 60 * time.Second      // Avoid excessive backoff

	// Retries fail fast on persistent loss

	DefaultRetries = 8
	MinRetries     = 1
	MaxRetries     = 20

	// AckDelay QUIC MAX_ACK_DELAY (RFC 9000 §13.2.1)

	DefaultAckDelay = 25 * time.Millisecond
	MinAckDelay     = 0

	// Buffer sizes

	DefaultRecvBufBytes = 1 << 20            // 1 MiB
	MinRecvBufBytes     = DefaultWindowBytes // Should be at least as large as window
	MaxRecvBufBytes     = 8 << 20            // 8 MiB
	DefaultSendBufBytes = 1 << 20
	MinSendBufBytes     = DefaultWindowBytes
	MaxSendBufBytes     = 8 << 20
)

// Config holds general settings for the UDP module.
type Config struct {
	ListenPort      int           `yaml:"listen_port,omitempty"` // Port to listen on for incoming connections (default 1791)
	PublicEndpoints []string      `yaml:"public_endpoints,omitempty"`
	DialTimeout     time.Duration `yaml:"dial_timeout,omitempty"` // Timeout for dialing connections (default 1 minute)

	FlowControl FlowControlConfig `yaml:"flow_control,omitempty"` // Flow control settings for UDP connections
}

// FlowControlConfig holds configuration for individual UDP connections.
type FlowControlConfig struct {
	MSS          int           // Maximum Segment Size (default 1187)
	WindowBytes  int           // Send window size in bytes (default 16 * MSS)
	RTO          time.Duration // Initial retransmission timeout (default 500ms)
	RTOMax       time.Duration // Maximum retransmission timeout (default 4s)
	RetryLimit   int           // Maximum retransmission attempts (default 8)
	IdleTimeout  time.Duration // Connection idle timeout (default 60s)
	AckDelay     time.Duration // Delayed ACK timer (default 25ms)
	RecvBufBytes int           // Receive buffer size (default 1MB)
	SendBufBytes int           // Send buffer size (default 1MB)
}

// Normalize sets sensible defaults for zero-values, clamps to safe ranges, and enforces invariants.
// See RFC 9000, RFC 8085, RFC 6298 for rationale.
func (c *FlowControlConfig) Normalize() {
	c.setDefaults()
	c.clampValues()
}

// setDefaults initializes zero-values with sensible defaults.
func (c *FlowControlConfig) setDefaults() {
	if c.MSS == 0 {
		c.MSS = DefaultMSS
	}
	if c.WindowBytes == 0 {
		c.WindowBytes = DefaultWindowBytes
	}
	if c.RTO == 0 {
		c.RTO = DefaultRTO
	}
	if c.RTOMax == 0 {
		c.RTOMax = DefaultRTOMax
	}
	if c.RetryLimit == 0 {
		c.RetryLimit = DefaultRetries
	}
	if c.AckDelay == 0 {
		c.AckDelay = DefaultAckDelay
	}
	if c.RecvBufBytes == 0 {
		c.RecvBufBytes = DefaultRecvBufBytes
	}
	if c.SendBufBytes == 0 {
		c.SendBufBytes = DefaultSendBufBytes
	}
}

// NOTE: normally i would not introduce such function but when it comes to
// parameters of network protocols,
// i believe it is better to keep things within certain range of values (
// all of which are stated at the top of this file)

// clampValues ensures all fields are within safe ranges and enforces invariants.
func (c *FlowControlConfig) clampValues() {
	c.MSS = clampInt(c.MSS, MinMSS, MaxMSS)
	c.WindowBytes = clampInt(c.WindowBytes, c.MSS, MaxWindowBytes)
	c.RTO = clampDur(c.RTO, MinRTO, MaxRTOCeiling)
	c.RTOMax = clampDur(c.RTOMax, c.RTO, MaxRTOCeiling)
	c.RetryLimit = clampInt(c.RetryLimit, MinRetries, MaxRetries)
	c.AckDelay = clampDur(c.AckDelay, MinAckDelay, c.RTO/2)
	c.RecvBufBytes = clampInt(c.RecvBufBytes, MinRecvBufBytes, MaxRecvBufBytes)
	c.SendBufBytes = clampInt(c.SendBufBytes, MinSendBufBytes, MaxSendBufBytes)
}

// clampInt clamps an integer value to a specified range.
func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// clampDur clamps a time.Duration value to a specified range.
func clampDur(v, lo, hi time.Duration) time.Duration {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

var defaultConfig = Config{
	ListenPort:  ListenPort,
	DialTimeout: time.Minute,
	FlowControl: FlowControlConfig{
		MSS:          DefaultMSS,
		WindowBytes:  DefaultWindowBytes,
		RTO:          DefaultRTO,
		RTOMax:       DefaultRTOMax,
		RetryLimit:   DefaultRetries,
		IdleTimeout:  60 * time.Second, // Default idle timeout of 1 minute
		AckDelay:     DefaultAckDelay,
		RecvBufBytes: DefaultRecvBufBytes,
		SendBufBytes: DefaultSendBufBytes,
	},
}

// RFC rationale summary:
//
// - MSS: QUIC requires 1200B UDP datagrams (RFC 9000 §14.1), clamped to avoid fragmentation (RFC 8085).
// - WindowBytes: conservative buffer, RFC 8085, must be >= MSS.
// - RTO/RTOMax: TCP discipline (RFC 6298), pragmatic for UDP, exponential backoff.
// - AckDelay: mirrors QUIC MAX_ACK_DELAY (RFC 9000 §13.2.1).
// - Buffer sizes: 1 MiB default, capped for safety, must be >= window.
// - All invariants enforced for safety and interoperability.
