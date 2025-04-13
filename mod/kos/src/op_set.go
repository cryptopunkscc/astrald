package kos

import (
	"bytes"
	"encoding"
	"encoding/base64"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opSetArgs struct {
	Key   string
	Type  string `query:"optional"`
	Raw   string `query:"optional"` // base64 encoded raw data
	Value string `query:"optional"` // text to unmarshal, must provide type
}

func (mod *Module) OpSet(ctx *astral.Context, q shell.Query, args opSetArgs) (err error) {
	var payload []byte

	switch {
	case len(args.Value) > 0:
		obj := mod.Objects.Blueprints().Make(args.Type)
		if obj == nil {
			return q.Reject()
		}

		m, ok := obj.(encoding.TextUnmarshaler)
		if !ok {
			return q.Reject()
		}

		err = m.UnmarshalText([]byte(args.Value))
		if err != nil {
			return q.Reject()
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

	conn, err := q.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()

	err = mod.db.Set(ctx.Identity(), args.Key, args.Type, payload)
	if err != nil {
		mod.log.Errorv(2, "errors setting %v:%v: %v", ctx.Identity(), args.Key, err)
		_, err = astral.Write(conn, astral.NewError(err.Error()), false)
		return err
	}

	_, err = astral.Write(conn, &astral.Ack{}, false)
	return err
}
