package relay

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	_data "github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/data"
	"github.com/cryptopunkscc/astrald/mod/relay"
)

func (mod *Module) DescribeData(ctx context.Context, dataID _data.ID, opts *data.DescribeOpts) []data.Descriptor {
	row, err := mod.dbFindByDataID(dataID)
	if err != nil {
		return nil
	}

	relayID, err := id.ParsePublicKeyHex(row.RelayID)
	if err != nil {
		return nil
	}

	targetID, err := id.ParsePublicKeyHex(row.TargetID)
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

	return []data.Descriptor{{
		Type: relay.CertDescriptorType,
		Data: relay.CertDescriptor{
			TargetID:      targetID,
			RelayID:       relayID,
			Direction:     relay.Direction(row.Direction),
			ExpiresAt:     row.ExpiresAt,
			ValidateError: verr,
		},
	}}
}
