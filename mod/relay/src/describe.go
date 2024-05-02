package relay

import (
	"context"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/mod/relay"
	"github.com/cryptopunkscc/astrald/object"
)

func (mod *Module) Describe(ctx context.Context, objectID object.ID, opts *desc.Opts) []*desc.Desc {
	var row dbCert
	err := mod.db.
		Where("data_id = ?", objectID).
		First(&row).Error
	if err != nil {
		return nil
	}

	var verr error
	cert, err := mod.LoadCert(objectID)
	if err != nil {
		verr = err
	} else {
		verr = cert.Validate()
	}

	return []*desc.Desc{{
		Source: mod.node.Identity(),
		Data: relay.CertDesc{
			TargetID:      row.TargetID,
			RelayID:       row.RelayID,
			Direction:     relay.Direction(row.Direction),
			ExpiresAt:     row.ExpiresAt,
			ValidateError: verr,
		},
	}}
}
