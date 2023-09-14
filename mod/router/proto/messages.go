package proto

import "github.com/cryptopunkscc/astrald/auth/id"

type QueryParams struct {
	Cert   []byte      `cslq:"[s]c"`
	Target id.Identity `cslq:"v"`
	Query  string      `cslq:"[c]c"`
}

type QueryResponse struct {
	Cert         []byte `cslq:"[s]c"`
	ProxyService string `cslq:"[c]c"`
}
