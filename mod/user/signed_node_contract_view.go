package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/log/styles"
	"github.com/cryptopunkscc/astrald/mod/log/views"
)

type SignedNodeContractView struct {
	*SignedNodeContract
}

func (view SignedNodeContractView) Render() string {
	return log.DefaultViewer.Render(log.Format(
		"Signed node contract (%v@%v) from %v to %v",
		view.UserID,
		view.NodeID,
		views.NewTimeViewStyled(&view.StartsAt, "2006-01-02 15:04:05.000", styles.DarkGreenText),
		views.NewTimeViewStyled(&view.ExpiresAt, "2006-01-02 15:04:05.000", styles.DarkGreenText),
	)...)
}

func init() {
	log.DefaultViewer.Set(SignedNodeContract{}.ObjectType(), func(object astral.Object) astral.Object {
		return &SignedNodeContractView{object.(*SignedNodeContract)}
	})
}
