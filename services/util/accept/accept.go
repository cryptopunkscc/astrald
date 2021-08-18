package accept

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/serializer"
)

func Request(
	ctx context.Context,
	request api.ConnectionRequest,
) (stream serializer.ReadWriteCloser) {
	stream = serializer.New(request.Accept())
	go func() {
		<-ctx.Done()
		_ = stream.Close()
	}()
	return stream
}
