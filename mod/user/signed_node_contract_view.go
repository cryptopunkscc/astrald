package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	log2 "github.com/cryptopunkscc/astrald/mod/log"
)

type SignedNodeContractView struct {
	*SignedNodeContract
}

func (view SignedNodeContractView) Render() string {
	return log.DefaultViewer.Render(log.Format(
		"Signed node contract (%v@%v) from %v to %v",
		view.UserID,
		view.NodeID,
		log2.NewTimeViewWithStyle(&view.StartsAt, "2006-01-02 15:04:05.000", log2.DarkGreenText),
		log2.NewTimeViewWithStyle(&view.ExpiresAt, "2006-01-02 15:04:05.000", log2.DarkGreenText),
	)...)
}

func init() {
	log.DefaultViewer.Set(SignedNodeContract{}.ObjectType(), func(object astral.Object) astral.Object {
		return &SignedNodeContractView{object.(*SignedNodeContract)}
	})
}
