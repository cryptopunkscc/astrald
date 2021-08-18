package serializer

import (
	"encoding/binary"
	"io"
)

type Reader struct {
	io.Reader
}

func (p *Reader) ReadByte() (byte, error) {
	buff, err := p.ReadN(1)
	if err != nil {
		return 0, err
	}
	return buff[0], nil
}

func (p *Reader) ReadUint8() (uint16, error) {
	buff, err := p.ReadN(1)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(buff[:]), nil
}

func (p *Reader) ReadUint16() (uint16, error) {
	buff, err := p.ReadN(2)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(buff[:]), nil
}

func (p *Reader) ReadUint32() (uint32, error) {
	buff, err := p.ReadN(4)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(buff[:]), nil
}

func (p *Reader) ReadUint64() (uint64, error) {
	buff, err := p.ReadN(8)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(buff[:]), nil
}


func (p *Reader) ReadN(n int) ([]byte, error) {
	buff := make([]byte, n)
	read, err := p.Read(buff)
	if err != nil {
		return nil, err
	}
	return buff[:read], nil
}

func (p *Reader) ReadWithSize8() (buff []byte, err error) {
	size, err := p.ReadUint8()
	if err != nil {
		return
	}
	buff, err = p.ReadN(int(size))
	if err != nil {
		return
	}
	return
}

func (p *Reader) ReadWithSize16() (buff []byte, err error) {
	size, err := p.ReadUint16()
	if err != nil {
		return
	}
	buff, err = p.ReadN(int(size))
	if err != nil {
		return
	}
	return
}


func (p *Reader) ReadWithSize32() (buff []byte, err error) {
	size, err := p.ReadUint32()
	if err != nil {
		return
	}
	buff, err = p.ReadN(int(size))
	if err != nil {
		return
	}
	return
}

func (p *Reader) ReadString(n int) (string, error) {
	buff, err := p.ReadN(n)
	if err != nil {
		return "", err
	}
	return string(buff[:]), nil
}

func (p *Reader) ReadStringWithSize8() (string, error) {
	buff, err := p.ReadWithSize8()
	if err != nil {
		return "", err
	}
	return string(buff), err
}

func (p *Reader) ReadStringWithSize16() (string, error) {
	buff, err := p.ReadWithSize16()
	if err != nil {
		return "", err
	}
	return string(buff), err
}

func (p *Reader) ReadStringWithSize32() (string, error) {
	buff, err := p.ReadWithSize32()
	if err != nil {
		return "", err
	}
	return string(buff), err
}
