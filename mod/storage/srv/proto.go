package proto

import (
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/storage"
)

type Request interface {
	Query() string
	Action() string
	Hidden() bool
}

type Response struct {
	Err string `json:"error,omitempty"`
}

func (e Response) Error() string { return e.Err }

func (r ReadAllReq) Query() string  { return "read" }
func (r ReadAllReq) Action() string { return storage.OpenAction }
func (r ReadAllReq) Hidden() bool   { return false }

type ReadAllReq struct {
	ID      data.ID `json:"id"`
	Offset  uint64  `json:"offset,omitempty"`
	Virtual bool    `json:"virtual,omitempty"`
	Network bool    `json:"network,omitempty"`
	Filter  string  `json:"filter,omitempty"`
}

type ReadAllResp struct {
	Response
	Bytes []byte `json:"bytes"`
}

func (r PutReq) Query() string  { return "put" }
func (r PutReq) Action() string { return storage.CreateAction }
func (r PutReq) Hidden() bool   { return false }

type PutReq struct {
	Bytes []byte `json:"bytes"`
	*storage.CreateOpts
}

type PutResp struct {
	Response
	ID *data.ID `json:"id"`
}

func (r OpenReq) Query() string  { return "open" }
func (r OpenReq) Action() string { return storage.OpenAction }
func (r OpenReq) Hidden() bool   { return true }

type OpenReq struct {
	ID      data.ID `json:"id"`
	Offset  uint64  `json:"offset,omitempty"`
	Virtual bool    `json:"virtual,omitempty"`
	Network bool    `json:"network,omitempty"`
	Filter  string  `json:"filter,omitempty"`
}

func (r CreateReq) Query() string  { return "create" }
func (r CreateReq) Action() string { return storage.CreateAction }
func (r CreateReq) Hidden() bool   { return true }

type CreateReq struct {
	*storage.CreateOpts
}

func (r PurgeReq) Query() string  { return "purge" }
func (r PurgeReq) Action() string { return storage.PurgeAction }
func (r PurgeReq) Hidden() bool   { return false }

type PurgeReq struct {
	ID data.ID `json:"id"`
	*storage.PurgeOpts
}

type PurgeResp struct {
	Response
	Total int `json:"total"`
}
