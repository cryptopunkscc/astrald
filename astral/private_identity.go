package astral

import (
	"bytes"
	"encoding/hex"
	"errors"
	"io"

	"github.com/btcsuite/btcd/btcec/v2"
)

type PrivateIdentity Identity

const privKeyBufSize = 32

// astral

func (PrivateIdentity) ObjectType() string {
	return "identity.secp256k1.private"
}

func (p PrivateIdentity) WriteTo(w io.Writer) (int64, error) {
	if p.privateKey == nil {
		return 0, errors.New("private key is nil")
	}

	m, err := w.Write(p.privateKey.Serialize())
	return int64(m), err
}

func (p *PrivateIdentity) ReadFrom(r io.Reader) (int64, error) {
	var buf = make([]byte, privKeyBufSize)
	n, err := r.Read(buf)
	if err != nil {
		return int64(n), err
	}

	p.privateKey, p.publicKey = btcec.PrivKeyFromBytes(buf)
	if p.privateKey == nil {
		return int64(n), errors.New("parse error")
	}

	return int64(n), err
}

// text

func (p PrivateIdentity) MarshalText() (text []byte, err error) {
	var buf = &bytes.Buffer{}
	_, err = p.WriteTo(buf)
	if err != nil {
		return
	}

	return []byte(hex.EncodeToString(buf.Bytes())), nil
}

func (p *PrivateIdentity) UnmarshalText(text []byte) error {
	buf, err := hex.DecodeString(string(text))
	if err != nil {
		return err
	}

	p.privateKey, p.publicKey = btcec.PrivKeyFromBytes(buf)
	if p.privateKey == nil {
		return errors.New("parse error")
	}

	return nil
}

// ...

func (p *PrivateIdentity) String() string {
	text, _ := p.MarshalText()
	return string(text)
}

func init() {
	if err := DefaultBlueprints.Add(&PrivateIdentity{}); err != nil {
		panic(err)
	}
}
