package relay

import (
	"context"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/mod/relay"
)

func (mod *Module) Describe(ctx context.Context, dataID data.ID, opts *desc.Opts) []*desc.Desc {
	var row dbCert
	err := mod.db.
		Where("data_id = ?", dataID).
		First(&row).Error
	if err != nil {
		return nil
	}

	var verr error
	cert, err := mod.LoadCert(dataID)
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
