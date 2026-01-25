package dir

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
)

func (client *Client) Filters(ctx *astral.Context) (filters []string, err error) {
	ch, err := client.queryCh(ctx, "dir.filters", nil)
	if err != nil {
		return nil, err
	}

	// collect the list of filters
	err = ch.Switch(
		channel.CollectStrings[*astral.String8](&filters),
		channel.StopOnEOS,
		channel.PassErrors,
	)
	if err != nil {
		return nil, err
	}

	return
}
