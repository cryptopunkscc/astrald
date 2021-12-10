package contacts

import (
	"bytes"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/enc"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/jxskiss/base62"
	"io"
	"strings"
)

const infoPrefix = "node1"

type Contact struct {
	Identity  id.Identity
	Alias     string
	Addresses []infra.Addr
}

func NewContact(identity id.Identity) *Contact {
	return &Contact{
		Identity:  identity,
		Addresses: make([]infra.Addr, 0),
	}
}

func (contact *Contact) Add(addr infra.Addr) {
	for _, a := range contact.Addresses {
		if infra.AddrEqual(a, addr) {
			return
		}
	}
	contact.Addresses = append(contact.Addresses, addr)
}

func (contact Contact) Pack() []byte {
	buf := &bytes.Buffer{}
	_ = writeInfo(buf, &contact)
	return buf.Bytes()
}

func (contact Contact) String() string {
	return infoPrefix + base62.EncodeToString(contact.Pack())
}

func ParseInfo(s string) (*Contact, error) {
	str := strings.TrimPrefix(s, infoPrefix)

	data, err := base62.DecodeString(str)
	if err != nil {
		return nil, err
	}

	return Unpack(data)
}

func Unpack(data []byte) (*Contact, error) {
	return readInfo(bytes.NewReader(data))
}

func writeInfo(w io.Writer, c *Contact) error {
	err := enc.WriteIdentity(w, c.Identity)
	if err != nil {
		return err
	}

	err = enc.WriteL8String(w, c.Alias)
	if err != nil {
		return err
	}

	addrs := c.Addresses[:]
	if len(addrs) > 255 {
		addrs = addrs[:255]
	}

	err = enc.Write(w, uint8(len(addrs)))
	if err != nil {
		return err
	}

	for _, addr := range addrs {
		if err := enc.WriteAddr(w, addr); err != nil {
			return nil
		}
	}

	return nil
}

func readInfo(r io.Reader) (*Contact, error) {
	_id, err := enc.ReadIdentity(r)
	if err != nil {
		return nil, err
	}

	alias, err := enc.ReadL8String(r)
	if err != nil {
		return nil, err
	}

	count, err := enc.ReadUint8(r)
	if err != nil {
		return nil, err
	}

	addrs := make([]infra.Addr, 0, count)
	for i := 0; i < int(count); i++ {
		addr, err := enc.ReadAddr(r)
		if err != nil {
			return nil, err
		}
		addrs = append(addrs, addr)
	}

	return &Contact{
		Identity:  _id,
		Alias:     alias,
		Addresses: addrs,
	}, nil
}
