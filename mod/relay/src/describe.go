package relay

import (
	"context"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/relay"
)

func (mod *Module) Describe(ctx context.Context, dataID data.ID, opts *content.DescribeOpts) []*content.Descriptor {
	row, err := mod.dbFindByDataID(dataID)
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

	return []*content.Descriptor{{
		Source: mod.node.Identity(),
		Data: relay.CertDescriptor{
			TargetID:      row.TargetID,
			RelayID:       row.RelayID,
			Direction:     relay.Direction(row.Direction),
			ExpiresAt:     row.ExpiresAt,
			ValidateError: verr,
		},
	}}
}
