package connect

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/serializer"
)

func Local(
	ctx context.Context,
	core api.Core,
	port string,
	request byte,
) (*serializer.ReadWriteCloser, error) {
	return Remote(ctx, core, "", port, request)
}

func Remote(
	ctx context.Context,
	core api.Core,
	identity api.Identity,
	port string,
	request byte,
) (*serializer.ReadWriteCloser, error) {
	stream, err := core.Network().Connect(identity, port)
	if err != nil {
		return nil, err
	}
	go func() {
		<-ctx.Done()
		_ = stream.Close()
	}()
	s := serializer.New(stream)
	err = s.WriteByte(request)
	if err != nil {
		return nil, err
	}
	return s, nil
}
