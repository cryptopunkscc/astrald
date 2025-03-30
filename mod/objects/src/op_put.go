package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"io"
)

type opPutArgs struct {
	Size int
}

func (mod *Module) OpPut(ctx *astral.Context, q shell.Query, args opPutArgs) (err error) {
	if !mod.Auth.Authorize(q.Caller(), objects.ActionWrite, nil) {
		return q.Reject()
	}

	create, err := mod.Create(&objects.CreateOpts{Alloc: args.Size})
	if err != nil {
		return q.Reject()
	}

	stream, err := shell.AcceptStream(q)
	if err != nil {
		return err
	}
	defer stream.Close()

	_, err = io.CopyN(create, stream, int64(args.Size))
	if err != nil {
		stream.Write([]byte{1})
		return
	}

	objectID, err := create.Commit()
	if err != nil {
		stream.Write([]byte{1})
		return
	}

	mod.log.Infov(2, "%v committed %v", q.Caller(), objectID)

	_, err = stream.Write([]byte{0})
	if err != nil {
		return
	}

	_, err = objectID.WriteTo(stream)
	return
}
