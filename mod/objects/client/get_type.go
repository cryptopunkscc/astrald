package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
)

// Deprecated: use Probe instead.
func (client *Client) GetType(ctx *astral.Context, objectID *astral.ObjectID) (typ string, err error) {
	ch, err := client.queryCh(ctx, "objects.get_type", query.Args{
		"id": objectID,
	})
	if err != nil {
		return "", err
	}
	defer ch.Close()

	err = ch.Switch(
		channel.ExpectString[*astral.String8](&typ),
		channel.PassErrors,
		channel.WithContext(ctx),
	)

	return
}
