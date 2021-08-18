package accept

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/serializer"
	"io"
)

func Request(
	ctx context.Context,
	request api.ConnectionRequest,
) (stream *serializer.ReadWriteCloser) {
	stream = serializer.New(request.Accept())
	go func(closer io.Closer) {
		<-ctx.Done()
		_ = closer.Close()
	}(stream.Closer)
	return stream
}