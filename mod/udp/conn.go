package udp

// DatagramWriter is how Conn sends bytes to its peer.
type DatagramWriter interface {
	WriteDatagram(b []byte) error
}

// DatagramReceiver is how Conn *receives* parsed packets when it does not own a socket read loop.
// (For active conns, the recvLoop calls HandleDatagram itself.)
type DatagramReceiver interface {
	HandleDatagram(raw []byte) // fast path: parse + process (ACK/data)
}
