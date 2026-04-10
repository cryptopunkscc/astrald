package auth

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	log2 "github.com/cryptopunkscc/astrald/mod/log"
)

type SignedContractView struct {
	*SignedContract
}

func (view SignedContractView) Render() string {
	return log.DefaultViewer.Render(log.Format(
		"Signed contract (%v -> %v) until %v",
		view.Issuer,
		view.Subject,
		log2.NewTimeViewWithStyle(&view.ExpiresAt, "2006-01-02 15:04:05.000", log2.DarkGreenText),
	)...)
}

func init() {
	log.DefaultViewer.Set(SignedContract{}.ObjectType(), func(object astral.Object) astral.Object {
		return &SignedContractView{object.(*SignedContract)}
	})
}
