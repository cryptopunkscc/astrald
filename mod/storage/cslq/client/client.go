package storage

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/lib/astral"
	"github.com/cryptopunkscc/astrald/mod/storage"
)

const (
	Port = "storage."
)

const (
	ReadAll = byte(iota) + 1
	Put
	AddOpener
	RemoveOpener
	AddCreator
	RemoveCreator
	AddPurger
	RemovePurger
	OpenerOpen
	CreatorCreate
	PurgerPurge
)

type Client struct {
	target id.Identity
	port   string
}

func NewClient(target id.Identity) *Client {
	return &Client{target: target, port: Port}
}

func (s Client) Port(port string) Client {
	s.port = port
	return s
}

func (s Client) Open(dataID data.ID, opts *storage.OpenOpts) (storage.Reader, error) {
	client := NewTargetClient(s.target, s.port).Route(OpenerOpen)
	return client.Open(dataID, opts)
}

func (s Client) Create(opts *storage.CreateOpts) (storage.Writer, error) {
	client := NewTargetClient(s.target, s.port).Route(CreatorCreate)
	return client.Create(opts)
}

func (s Client) Purge(dataID data.ID, opts *storage.PurgeOpts) (int, error) {
	client := NewTargetClient(s.target, s.port).Route(PurgerPurge)
	return client.Purge(dataID, opts)
}

func (s Client) ReadAll(id data.ID, opts *storage.OpenOpts) (b []byte, err error) {
	conn, err := s.query(ReadAll, "v v", id, opts)
	if err != nil {
		return
	}
	defer conn.Close()
	if err = cslq.Decode(conn, "[l]c", &b); err != nil {
		return
	}
	return
}

func (s Client) Put(b []byte, opts *storage.CreateOpts) (i data.ID, err error) {
	conn, err := s.query(Put, "[l]c v", b, opts)
	if err != nil {
		return
	}
	defer conn.Close()
	if err = cslq.Decode(conn, "v", &i); err != nil {
		return
	}
	return
}

func (s Client) AddOpener(name string, opener storage.Opener, priority int) (err error) {
	port := opener.(fmt.Stringer).String()
	return s.add(AddOpener, name, "[c]c l", port, priority)
}

func (s Client) RemoveOpener(name string) (err error) {
	return s.remove(RemoveOpener, name)
}

func (s Client) AddCreator(name string, creator storage.Creator, priority int) (err error) {
	port := creator.(fmt.Stringer).String()
	return s.add(AddCreator, name, "[c]c l", port, priority)
}

func (s Client) RemoveCreator(name string) (err error) {
	return s.remove(RemoveCreator, name)
}

func (s Client) AddPurger(name string, purger storage.Purger) error {
	return s.add(AddPurger, name, "[c]c", purger)
}

func (s Client) RemovePurger(name string) error {
	return s.remove(RemovePurger, name)
}

func (s Client) add(cmd byte, name string, format string, args ...any) (err error) {
	args = append([]any{name}, args...)
	conn, err := s.query(cmd, "[c]c"+format, args...)
	if err != nil {
		return
	}
	defer conn.Close()
	return
}

func (s Client) remove(cmd byte, name string) (err error) {
	conn, err := s.query(cmd, "[c]c", name)
	if err != nil {
		return
	}
	defer conn.Close()
	return
}

func (s Client) query(cmd byte, format string, args ...any) (conn *astral.Conn, err error) {
	return query(s.target, s.port, cmd, format, args)
}
