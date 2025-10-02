package rudp

import "time"

// Transport default constants (exported for visibility in tests & integration)
const (
	DefaultMSS          = 1200 - 13 // 1187 (1200 minus header)
	DefaultWndPkts      = 1024
	DefaultRTO          = 200 * time.Millisecond
	DefaultRTOMax       = 4 * time.Second
	DefaultRetries      = 8
	DefaultAckDelay     = 5 * time.Millisecond
	DefaultRecvBufBytes = 16 << 20
	DefaultSendBufBytes = 16 << 20
)

// Config holds reliability / buffering parameters for the rudp transport.
type Config struct {
	MaxSegmentSize            int           `yaml:"max_segment_size"`
	MaxWindowPackets          int           `yaml:"max_window_packets"`
	RetransmissionInterval    time.Duration `yaml:"retransmission_interval"`
	MaxRetransmissionInterval time.Duration `yaml:"max_retransmission_interval"`
	RetransmissionLimit       int           `yaml:"retransmission_limit"`
	AckDelay                  time.Duration `yaml:"ack_delay"`
	RecvBufBytes              int           `yaml:"recv_buf_bytes"`
	SendBufBytes              int           `yaml:"send_buf_bytes"`
}

// Normalize applies defaults to zero-value fields (no clamping beyond basic sanity here).
func (c *Config) Normalize() {
	if c.MaxSegmentSize == 0 {
		c.MaxSegmentSize = DefaultMSS
	}
	if c.MaxWindowPackets == 0 {
		c.MaxWindowPackets = DefaultWndPkts
	}
	if c.RetransmissionInterval == 0 {
		c.RetransmissionInterval = DefaultRTO
	}
	if c.MaxRetransmissionInterval == 0 {
		c.MaxRetransmissionInterval = DefaultRTOMax
	}
	if c.RetransmissionLimit == 0 {
		c.RetransmissionLimit = DefaultRetries
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

// - AckDelay: mirrors QUIC MAX_ACK_DELAY (RFC 9000 §13.2.1).
// - Buffer sizes: 1 MiB default, capped for safety, must be >= aggregate window.
// - Aggregate window bytes can be derived as MaxWindowPackets * MaxSegmentSize.
