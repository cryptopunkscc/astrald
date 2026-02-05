package apphost

import (
	"crypto/sha256"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/streams"
	"io"
)

type AppContract struct {
	AppID     *astral.Identity
	HostID    *astral.Identity
	StartsAt  astral.Time
	ExpiresAt astral.Time
	AppSig    astral.Bytes8
	HostSig   astral.Bytes8
}

var _ astral.Object = &AppContract{}

func (a AppContract) ObjectType() string { return "mod.apphost.app_contract" }

func (a AppContract) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(a).WriteTo(w)
}

func (a *AppContract) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(a).ReadFrom(r)
}

func (a *AppContract) SignableHash() []byte {
	var hash = sha256.New()
	_, err := streams.WriteAllTo(hash,
		astral.String8(a.ObjectType()),
		a.AppID,
		a.HostID,
		a.StartsAt,
		a.ExpiresAt,
	)
	if err != nil {
		return nil
	}
	return hash.Sum(nil)
}

func init() {
	astral.Add(&AppContract{})
}
