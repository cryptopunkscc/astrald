package connect

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/serialize"
)

func Local(
	ctx context.Context,
	core api.Core,
	port string,
	request byte,
) (*serialize.Serializer, error) {
	return Remote(ctx, core, "", port, request)
}

func Remote(
	ctx context.Context,
	core api.Core,
	identity api.Identity,
	port string,
	request byte,
) (*serialize.Serializer, error) {
	stream, err := core.Network().Connect(identity, port)
	if err != nil {
		return nil, err
	}
	go func() {
		<-ctx.Done()
		_ = stream.Close()
	}()
	s := serialize.NewSerializer(stream)
	err = s.WriteByte(request)
	if err != nil {
		return nil, err
	}
	return &s, nil
}
