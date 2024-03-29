package brontide

import (
	"bytes"
	"github.com/btcsuite/btcd/btcec/v2"
	"io"
	"math"
)

// Conn is an implementation of net.Conn which enforces an authenticated key
// exchange and message encryption protocol dubbed "Brontide" after initial TCP
// connection establishment. In the case of a successful handshake, all
// messages sent via the .Write() method are encrypted with an AEAD cipher
// along with an encrypted length-prefix. See the Machine struct for
// additional details w.r.t to the handshake and encryption scheme.
type Conn struct {
	conn    io.ReadWriteCloser
	noise   *Machine
	readBuf bytes.Buffer
}

// ReadNextMessage uses the connection in a message-oriented manner, instructing
// it to read the next _full_ message with the brontide stream. This function
// will block until the read of the header and body succeeds.
//
// NOTE: This method SHOULD NOT be used in the case that the connection may be
// adversarial and induce long delays. If the caller needs to set read deadlines
// appropriately, it is preferred that they use the split ReadNextHeader and
// ReadNextBody methods so that the deadlines can be set appropriately on each.
func (c *Conn) ReadNextMessage() ([]byte, error) {
	return c.noise.ReadMessage(c.conn)
}

// ReadNextHeader uses the connection to read the next header from the brontide
// stream. This function will block until the read of the header succeeds and
// return the packet length (including MAC overhead) that is expected from the
// subsequent call to ReadNextBody.
func (c *Conn) ReadNextHeader() (uint32, error) {
	return c.noise.ReadHeader(c.conn)
}

// ReadNextBody uses the connection to read the next message body from the
// brontide stream. This function will block until the read of the body succeeds
// and return the decrypted payload. The provided buffer MUST be the packet
// length returned by the preceding call to ReadNextHeader.
func (c *Conn) ReadNextBody(buf []byte) ([]byte, error) {
	return c.noise.ReadBody(c.conn, buf)
}

// Read reads data from the connection.  Read can be made to time out and
// return an Error with Timeout() == true after a fixed time limit; see
// SetDeadline and SetReadDeadline.
//
// Part of the net.Conn interface.
func (c *Conn) Read(b []byte) (n int, err error) {
	// In order to reconcile the differences between the record abstraction
	// of our AEAD connection, and the stream abstraction of TCP, we
	// maintain an intermediate read buffer. If this buffer becomes
	// depleted, then we read the next record, and feed it into the
	// buffer. Otherwise, we read directly from the buffer.
	if c.readBuf.Len() == 0 {
		plaintext, err := c.noise.ReadMessage(c.conn)
		if err != nil {
			return 0, err
		}

		if _, err := c.readBuf.Write(plaintext); err != nil {
			return 0, err
		}
	}

	return c.readBuf.Read(b)
}

// Write writes data to the connection.  Write can be made to time out and
// return an Error with Timeout() == true after a fixed time limit; see
// SetDeadline and SetWriteDeadline.
//
// Part of the net.Conn interface.
func (c *Conn) Write(b []byte) (n int, err error) {
	// If the message doesn't require any chunking, then we can go ahead
	// with a single write.
	if len(b) <= math.MaxUint16 {
		err = c.noise.WriteMessage(b)
		if err != nil {
			return 0, err
		}
		return c.noise.Flush(c.conn)
	}

	// If we need to split the message into fragments, then we'll write
	// chunks which maximize usage of the available payload.
	chunkSize := math.MaxUint16

	bytesToWrite := len(b)
	bytesWritten := 0
	for bytesWritten < bytesToWrite {
		// If we're on the last chunk, then truncate the chunk size as
		// necessary to avoid an out-of-bounds array memory access.
		if bytesWritten+chunkSize > len(b) {
			chunkSize = len(b) - bytesWritten
		}

		// Slice off the next chunk to be written based on our running
		// counter and next chunk size.
		chunk := b[bytesWritten : bytesWritten+chunkSize]
		if err := c.noise.WriteMessage(chunk); err != nil {
			return bytesWritten, err
		}

		n, err := c.noise.Flush(c.conn)
		bytesWritten += n
		if err != nil {
			return bytesWritten, err
		}
	}

	return bytesWritten, nil
}

// WriteMessage encrypts and buffers the next message p for the connection. The
// ciphertext of the message is prepended with an encrypt+auth'd length which
// must be used as the AD to the AEAD construction when being decrypted by the
// other side.
//
// NOTE: This DOES NOT write the message to the wire, it should be followed by a
// call to Flush to ensure the message is written.
func (c *Conn) WriteMessage(b []byte) error {
	return c.noise.WriteMessage(b)
}

// Flush attempts to write a message buffered using WriteMessage to the
// underlying connection. If no buffered message exists, this will result in a
// NOP. Otherwise, it will continue to write the remaining bytes, picking up
// where the byte stream left off in the event of a partial write. The number of
// bytes returned reflects the number of plaintext bytes in the payload, and
// does not account for the overhead of the header or MACs.
//
// NOTE: It is safe to call this method again iff a timeout error is returned.
func (c *Conn) Flush() (int, error) {
	return c.noise.Flush(c.conn)
}

// Close closes the connection. Any blocked Read or Write operations will be
// unblocked and return errors.
func (c *Conn) Close() error {
	return c.conn.Close()
}

// RemotePub returns the remote peer's static public key.
func (c *Conn) RemotePub() *btcec.PublicKey {
	return c.noise.remoteStatic
}

// LocalPub returns the local peer's static public key.
func (c *Conn) LocalPub() *btcec.PublicKey {
	return c.noise.localStatic.PubKey()
}
