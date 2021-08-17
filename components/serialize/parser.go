package serialize

import (
	"encoding/binary"
	"io"
)

type Parser struct {
	io.Reader
}

func NewParser(reader io.Reader) *Parser {
	return &Parser{Reader: reader}
}

func (p *Parser) ReadWithSize() (buff []byte, err error) {
	size, err := p.ReadSize()
	if err != nil {
		return
	}
	buff, err = p.ReadN(size)
	if err != nil {
		return
	}
	return
}

func (p *Parser) ReadStringWithSize() (string, error) {
	buff, err := p.ReadWithSize()
	if err != nil {
		return "", err
	}
	return string(buff), err
}

func (p *Parser) ReadN(n int) ([]byte, error) {
	buff := make([]byte, n)
	read, err := p.Read(buff)
	if err != nil {
		return nil, err
	}

	return buff[:read], nil
}

func (p *Parser) ReadByte() (byte, error) {
	buff, err := p.ReadN(1)
	if err != nil {
		return 0, err
	}
	return buff[0], nil
}

func (p *Parser) ReadUint16() (uint16, error) {
	buff, err := p.ReadN(2)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(buff[:]), nil
}

func (p *Parser) ReadSize() (int, error) {
	return p.ReadInt()
}

func (p *Parser) ReadInt() (int, error) {
	i, err := p.ReadUint32()
	return int(i), err
}

func (p *Parser) ReadUint32() (uint32, error) {
	buff, err := p.ReadN(4)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(buff[:]), nil
}

func (p *Parser) ReadUint64() (uint64, error) {
	buff, err := p.ReadN(8)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(buff[:]), nil
}

func (p *Parser) ReadString(n int) (string, error) {
	buff, err := p.ReadN(n)
	if err != nil {
		return "", err
	}
	return string(buff[:]), nil
}
