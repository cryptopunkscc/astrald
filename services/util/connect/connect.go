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
) (sio.ReadWriteCloser, error) {
	return Remote(ctx, core, "", port)
}

func Remote(
	ctx context.Context,
	core api.Core,
	identity api.Identity,
	port string,
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
	return s, nil
}

func LocalRequest(
	ctx context.Context,
	core api.Core,
	port string,
	request byte,
) (sio.ReadWriteCloser, error) {
	return RemoteRequest(ctx, core, "", port, request)
}

func RemoteRequest(
	ctx context.Context,
	core api.Core,
	identity api.Identity,
	port string,
	request byte,
) (sio.ReadWriteCloser, error) {
	s, err := Remote(ctx, core, identity, port)
	if err != nil {
		return nil, err
	}
	err = s.WriteByte(request)
	if err != nil {
		return nil, err
	}
	return s, nil
}
