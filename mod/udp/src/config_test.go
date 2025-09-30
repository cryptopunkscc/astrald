package udp

import (
	"testing"
	"time"
)

func TestFlowControlConfigDefaults(t *testing.T) {
	def := defaultConfig.FlowControl
	if def.MSS != DefaultMSS || def.WindowBytes != DefaultWindowBytes || def.RTO != DefaultRTO || def.RTOMax != DefaultRTOMax || def.RetryLimit != DefaultRetries || def.AckDelay != DefaultAckDelay || def.RecvBufBytes != DefaultRecvBufBytes || def.SendBufBytes != DefaultSendBufBytes {
		t.Errorf("defaultConfig.FlowControl does not match expected defaults: %+v", def)
	}
}

func TestFlowControlConfigClamp(t *testing.T) {
	tests := []struct {
		name     string
		input    ReliableTransportConfig
		expected ReliableTransportConfig
	}{
		{
			name: "Values below range are clamped",
			input: ReliableTransportConfig{
				MSS:          100,
				WindowBytes:  100,
				RTO:          5 * time.Millisecond,
				RTOMax:       5 * time.Millisecond,
				RetryLimit:   0,
				AckDelay:     -1 * time.Millisecond,
				RecvBufBytes: 100,
				SendBufBytes: 100,
			},
			expected: ReliableTransportConfig{
				MSS:          MinMSS,
				WindowBytes:  MinMSS,
				RTO:          MinRTO,
				RTOMax:       MinRTO,
				RetryLimit:   MinRetries,
				AckDelay:     MinAckDelay,
				RecvBufBytes: MinRecvBufBytes,
				SendBufBytes: MinSendBufBytes,
			},
		},
		{
			name: "Values above range are clamped",
			input: ReliableTransportConfig{
				MSS:          2000,
				WindowBytes:  2 << 20,
				RTO:          70 * time.Second,
				RTOMax:       70 * time.Second,
				RetryLimit:   50,
				AckDelay:     1 * time.Second,
				RecvBufBytes: 16 << 20,
				SendBufBytes: 16 << 20,
			},
			expected: ReliableTransportConfig{
				MSS:          MaxMSS,
				WindowBytes:  MaxWindowBytes,
				RTO:          MaxRTOCeiling,
				RTOMax:       MaxRTOCeiling,
				RetryLimit:   MaxRetries,
				AckDelay:     1 * time.Second, // Correct expected value
				RecvBufBytes: MaxRecvBufBytes,
				SendBufBytes: MaxSendBufBytes,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := test.input
			input.clampValues()
			if input != test.expected {
				t.Errorf("clampValues() failed.\nGot: %+v\nExpected: %+v", input, test.expected)
			}
		})
	}
}
