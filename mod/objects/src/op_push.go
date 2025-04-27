package objects

import (
	"bytes"
	"encoding/binary"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/object"
	"io"
)

const maxPushSize = 4096

type opPushArgs struct {
	Size int
}

func (mod *Module) OpPush(ctx *astral.Context, q shell.Query, args opPushArgs) (err error) {
	if args.Size > maxPushSize {
		return q.Reject()
	}

	stream := q.Accept()
	defer stream.Close()

	var buf = make([]byte, args.Size)
	_, err = io.ReadFull(stream, buf)
	if err != nil {
		mod.log.Errorv(1, "%v push read error: %v", q.Caller(), err)
		binary.Write(stream, binary.BigEndian, false)
		return
	}

	obj, _, err := mod.Blueprints().Read(bytes.NewReader(buf), true)
	if err != nil {
		mod.log.Errorv(1, "%v push read object error: %v", q.Caller(), err)
		binary.Write(stream, binary.BigEndian, false)
		return
	}

	var objectID *object.ID
	objectID, err = astral.ResolveObjectID(obj)
	if err != nil {
		return
	}

	var push = &objects.SourcedObject{
		Source: q.Caller(),
		Object: obj,
	}

	if !mod.receive(push) {
		mod.log.Errorv(1, "rejected %v from %v (%v)", obj.ObjectType(), q.Caller(), objectID)
		binary.Write(stream, binary.BigEndian, false)
		return
	}

	binary.Write(stream, binary.BigEndian, true)

	mod.log.Infov(1, "accepted %v from %v (%v)", obj.ObjectType(), q.Caller(), objectID)

	return
}
