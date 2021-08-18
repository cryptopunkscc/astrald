package accept

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/serialize"
)

func Request(
	ctx context.Context,
	request api.ConnectionRequest,
) (stream serialize.Serializer) {
	stream = serialize.NewSerializer(request.Accept())
	go func() {
		<-ctx.Done()
		_ = stream.Close()
	}()
	return stream
}