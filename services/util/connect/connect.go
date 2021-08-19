package connect

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/sio"
)

func Local(
	ctx context.Context,
	core api.Core,
	port string,
	request uint16,
) (sio.ReadWriteCloser, error) {
	return Remote(ctx, core, "", port, request)
}

func Remote(
	ctx context.Context,
	core api.Core,
	identity api.Identity,
	port string,
	request uint16,
) (sio.ReadWriteCloser, error) {
	stream, err := core.Network().Connect(identity, port)
	if err != nil {
		return nil, err
	}
	go func() {
		<-ctx.Done()
		_ = stream.Close()
	}()
	s := sio.New(stream)
	_, err = s.WriteUInt16(request)
	if err != nil {
		return nil, err
	}
	return s, nil
}
