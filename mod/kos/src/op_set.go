package kos

import (
	"bytes"
	"encoding"
	"encoding/base64"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
)

type opSetArgs struct {
	Key   string
	Type  string `query:"optional"`
	Raw   string `query:"optional"` // base64 encoded raw data
	Value string `query:"optional"` // text to unmarshal, must provide type
	Out   string `query:"optional"`
}

func (mod *Module) OpSet(ctx *astral.Context, q *ops.Query, args opSetArgs) (err error) {
	var payload []byte

	switch {
	case len(args.Value) > 0:
		obj := astral.New(args.Type)
		if obj == nil {
			return q.RejectWithCode(8)
		}

		m, ok := obj.(encoding.TextUnmarshaler)
		if !ok {
			return q.RejectWithCode(astral.CodeInternalError)
		}

		err = m.UnmarshalText([]byte(args.Value))
		if err != nil {
			return q.RejectWithCode(astral.CodeInternalError)
		}

		var buf = &bytes.Buffer{}
		_, err = obj.WriteTo(buf)
		if err != nil {
			return q.Reject()
		}

		payload = buf.Bytes()

	case len(args.Raw) > 0:
		payload, err = base64.StdEncoding.DecodeString(args.Raw)
		if err != nil {
			return q.Reject()
		}

	default:
		return q.Reject()
	}

	ch := channel.New(q.Accept(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	err = mod.db.Set(ctx.Identity(), args.Key, args.Type, payload)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	return ch.Send(&astral.Ack{})
}
