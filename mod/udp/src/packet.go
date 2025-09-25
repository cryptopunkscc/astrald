package udp

import (
	"bytes"
	"encoding/binary"

	"github.com/cryptopunkscc/astrald/mod/udp"
)

const (
	FlagSYN = 1 << 0
	FlagACK = 1 << 1
	FlagFIN = 1 << 2
)

// Packet represents a src UDP packet with TCP-like header
type Packet struct {
	Seq     uint32 // Sequence number (first byte seq of this segment)
	Ack     uint32 // Acknowledgment number (cumulative ack: all bytes < Ack received)
	Flags   uint8  // Bit flags: SYN, ACK, FIN, etc.
	Win     uint16 // Advertised receive window in BYTES
	Len     uint16 // Payload length
	Payload []byte // Actual data
}

// Marshal serializes the Packet into bytes for transmission
// Format: [Seq:4][Ack:4][Flags:1][Win:2][Len:2][Payload:N]
func (p *Packet) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, p.Seq); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, p.Ack); err != nil {
		return nil, err
	}
	if err := buf.WriteByte(p.Flags); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, p.Win); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, p.Len); err != nil {
		return nil, err
	}
	if _, err := buf.Write(p.Payload); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Unmarshal parses bytes into the Packet struct fields
func (p *Packet) Unmarshal(data []byte) error {
	if len(data) < 13 { // 4+4+1+2+2 = 13 bytes header
		return udp.ErrPacketTooShort
	}

	p.Seq = binary.BigEndian.Uint32(data[0:4])
	p.Ack = binary.BigEndian.Uint32(data[4:8])
	p.Flags = data[8]
	p.Win = binary.BigEndian.Uint16(data[9:11])
	p.Len = binary.BigEndian.Uint16(data[11:13])

	// Validate payload length
	if int(p.Len) != len(data)-13 {
		return udp.ErrInvalidPayloadLength
	}

	p.Payload = data[13:]
	return nil
}

// UnmarshalPacket parses bytes into a Packet instance.
func UnmarshalPacket(data []byte) (*Packet, error) {
	if len(data) < 13 { // Minimum header size: 4+4+1+2+2
		return nil, udp.ErrMalformedPacket
	}

	pkt := &Packet{}
	r := bytes.NewReader(data)
	if err := binary.Read(r, binary.BigEndian, &pkt.Seq); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.BigEndian, &pkt.Ack); err != nil {
		return nil, err
	}
	flags, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	pkt.Flags = flags
	if err := binary.Read(r, binary.BigEndian, &pkt.Win); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.BigEndian, &pkt.Len); err != nil {
		return nil, err
	}
	if int(pkt.Len) > len(data)-13 {
		return nil, udp.ErrMalformedPacket
	}
	pkt.Payload = make([]byte, pkt.Len)
	if _, err := r.Read(pkt.Payload); err != nil {
		return nil, err
	}

	return pkt, nil
}
