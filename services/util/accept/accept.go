package accept

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/sio"
)

func Request(
	ctx context.Context,
	request api.ConnectionRequest,
) (stream sio.ReadWriteCloser) {
	stream = sio.New(request.Accept())
	go func() {
		<-ctx.Done()
		_ = stream.Close()
	}()
	return stream
}
