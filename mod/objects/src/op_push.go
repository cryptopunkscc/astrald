package objects

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

const maxPushSize = 32 * 1024

type opPushArgs struct {
	Size int    `query:"optional"`
	In   string `query:"optional"`
	Out  string `query:"optional"`
}

func (mod *Module) OpPush(ctx *astral.Context, q shell.Query, args opPushArgs) (err error) {
	switch {
	case args.Size > maxPushSize:
		return q.Reject()
	case args.Size == 0:
		return mod.opPushStream(ctx, q, args)
	default:
		return mod.opPushSingle(ctx, q, args)
	}
}

func (mod *Module) opPushStream(ctx *astral.Context, q shell.Query, args opPushArgs) (err error) {
	ch := astral.NewChannelFmt(q.Accept(), args.In, args.Out)
	defer ch.Close()

	var acccepted, rejected int

	defer func() {
		mod.log.Logv(1, "push from %v: accepted %v, rejected %v", q.Caller(), acccepted, rejected)
	}()

	for {
		object, err := ch.Read()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				mod.log.Errorv(1, "op_push: channel read: %v", err)
			}
			return nil
		}

		if mod.receive(q.Caller(), object) {
			acccepted++
			err = ch.Write(&astral.Ack{})
		} else {
			rejected++
			err = ch.Write(astral.NewError("rejected"))
		}
		if err != nil {
			mod.log.Errorv(1, "op_push: channel write: %v", err)
			return nil
		}
	}
}

func (mod *Module) opPushSingle(ctx *astral.Context, q shell.Query, args opPushArgs) (err error) {
	stream := q.Accept()
	defer stream.Close()

	var buf = make([]byte, args.Size)
	_, err = io.ReadFull(stream, buf)
	if err != nil {
		mod.log.Errorv(1, "%v push read error: %v", q.Caller(), err)
		binary.Write(stream, binary.BigEndian, false)
		return
	}

	obj, _, err := mod.Blueprints().ReadCanonical(bytes.NewReader(buf))
	if err != nil {
		mod.log.Errorv(1, "%v push read object error: %v", q.Caller(), err)
		binary.Write(stream, binary.BigEndian, false)
		return
	}

	var objectID *astral.ObjectID
	objectID, err = astral.ResolveObjectID(obj)
	if err != nil {
		return
	}

	if !mod.receive(q.Caller(), obj) {
		mod.log.Errorv(1, "rejected %v from %v (%v)", obj.ObjectType(), q.Caller(), objectID)
		binary.Write(stream, binary.BigEndian, false)
		return
	}

	binary.Write(stream, binary.BigEndian, true)

	mod.log.Infov(1, "accepted %v from %v (%v)", obj.ObjectType(), q.Caller(), objectID)

	return
}
