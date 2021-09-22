package brontide

import (
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"io"
)

// PassiveHandshake performs the brontide handshake over provided transport as the responder.
func PassiveHandshake(conn io.ReadWriteCloser, localStatic *btcec.PrivateKey) (*Conn, error) {
	ecdh := &PrivKeyECDH{PrivKey: localStatic}

	c := &Conn{
		conn:  conn,
		noise: NewBrontideMachine(false, ecdh, nil),
	}

	var actOne [ActOneSize]byte
	if _, err := io.ReadFull(conn, actOne[:]); err != nil {
		c.conn.Close()
		return nil, rejectedConnErr(err, "")
	}
	if err := c.noise.RecvActOne(actOne); err != nil {
		c.conn.Close()
		return nil, rejectedConnErr(err, "")
	}

	actTwo, err := c.noise.GenActTwo()
	if err != nil {
		c.conn.Close()
		return nil, rejectedConnErr(err, "")
	}
	if _, err := conn.Write(actTwo[:]); err != nil {
		c.conn.Close()
		return nil, rejectedConnErr(err, "")
	}

	var actThree [ActThreeSize]byte
	if _, err := io.ReadFull(conn, actThree[:]); err != nil {
		c.conn.Close()
		return nil, rejectedConnErr(err, "")
	}
	if err := c.noise.RecvActThree(actThree); err != nil {
		c.conn.Close()
		return nil, rejectedConnErr(err, "")
	}

	return c, nil
}

// rejectedConnErr is a helper function that prepends the remote address of the
// failed connection attempt to the original error message.
func rejectedConnErr(err error, remoteAddr string) error {
	return fmt.Errorf("unable to accept connection from %v: %v", remoteAddr,
		err)
}
