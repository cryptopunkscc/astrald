package brontide

import (
	"github.com/btcsuite/btcd/btcec"
	"io"
)

// ActiveHandshake performs the brontide handshake over provided transport as the initiator.
func ActiveHandshake(conn io.ReadWriteCloser, localKey *btcec.PrivateKey, remoteKey *btcec.PublicKey) (*Conn, error) {
	b := &Conn{
		conn:  conn,
		noise: NewBrontideMachine(true, &PrivKeyECDH{localKey}, remoteKey),
	}

	actOne, err := b.noise.GenActOne()
	if err != nil {
		b.conn.Close()
		return nil, err
	}
	if _, err := conn.Write(actOne[:]); err != nil {
		b.conn.Close()
		return nil, err
	}

	var actTwo [ActTwoSize]byte
	if _, err := io.ReadFull(conn, actTwo[:]); err != nil {
		b.conn.Close()
		return nil, err
	}
	if err := b.noise.RecvActTwo(actTwo); err != nil {
		b.conn.Close()
		return nil, err
	}

	actThree, err := b.noise.GenActThree()
	if err != nil {
		b.conn.Close()
		return nil, err
	}
	if _, err := conn.Write(actThree[:]); err != nil {
		b.conn.Close()
		return nil, err
	}

	return b, nil
}
