package udp

import (
	"bytes"
	"testing"
)

func TestPacketMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name      string
		packet    Packet
		expectErr bool
	}{
		{
			name: "Valid Packet",
			packet: Packet{
				Seq:     1,
				Ack:     2,
				Flags:   0x01,
				Win:     1024,
				Len:     5,
				Payload: []byte("hello"),
			},
			expectErr: false,
		},
		{
			name: "Empty Payload",
			packet: Packet{
				Seq:     10,
				Ack:     20,
				Flags:   0x02,
				Win:     2048,
				Len:     0,
				Payload: []byte{},
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Marshal
			data, err := tt.packet.Marshal()
			if (err != nil) != tt.expectErr {
				t.Fatalf("Marshal error = %v, expectErr = %v", err, tt.expectErr)
			}

			// Test Unmarshal
			var unmarshaled Packet
			err = unmarshaled.Unmarshal(data)
			if (err != nil) != tt.expectErr {
				t.Fatalf("Unmarshal error = %v, expectErr = %v", err, tt.expectErr)
			}

			// Verify the unmarshaled packet matches the original
			if !bytes.Equal(unmarshaled.Payload, tt.packet.Payload) ||
				unmarshaled.Seq != tt.packet.Seq ||
				unmarshaled.Ack != tt.packet.Ack ||
				unmarshaled.Flags != tt.packet.Flags ||
				unmarshaled.Win != tt.packet.Win ||
				unmarshaled.Len != tt.packet.Len {
				t.Errorf("Unmarshaled packet does not match original. Got %+v, want %+v", unmarshaled, tt.packet)
			}
		})
	}
}

func TestUnmarshalPacket(t *testing.T) {
	data := []byte("testdata")
	_, err := UnmarshalPacket(data)
	if err == nil {
		t.Error("Expected error for invalid packet data, got nil")
	}
}
