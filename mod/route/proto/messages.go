package proto

import "github.com/cryptopunkscc/astrald/auth/id"

const (
	CmdCert  = "cert"
	CmdQuery = "query"
)

type Cmd struct {
	Cmd string `cslq:"[c]c"`
}

type QueryParams struct {
	Target id.Identity `cslq:"v"`
	Query  string      `cslq:"[c]c"`
}
