package srv

import (
	proto "github.com/cryptopunkscc/astrald/mod/storage/srv"
	"io"
	"net/url"
)

type Request interface {
	Query() string
	Action() string
	Hidden() bool
}

type Handlers[Context any] map[string]Handler[Context]

type Handler[Context any] struct {
	proto.Request
	handle HandleFunc[Context]
}

type HandleFunc[Context any] func(
	Context,
	io.ReadWriteCloser,
	UnmarshalFunc,
	Encoder,
	string,
) error

type UnmarshalFunc func([]byte, any) error

type Encoder interface {
	Encode(data any) error
}

func Bind[Context any, R Request, Response any](
	s Handlers[Context],
	request R,
	handle func(Context, R) (Response, error),
) {
	s[request.Query()] = Handler[Context]{
		Request: request,
		handle: func(
			ctx Context,
			conn io.ReadWriteCloser,
			decodeReq UnmarshalFunc,
			encodeResp Encoder,
			query string,
		) (err error) {
			u, err := url.Parse(query)
			if err != nil {
				return
			}

			req := request
			if err = decodeReq([]byte(u.RawQuery), &req); err != nil {
				return
			}

			var resp any
			if resp, err = handle(ctx, req); err != nil {
				resp = proto.Response{Err: err.Error()}
			}
			err = encodeResp.Encode(resp)
			return
		},
	}
}
