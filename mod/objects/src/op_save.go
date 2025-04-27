package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"io"
)

type opSaveArgs struct {
	Size   int
	Type   string `query:"optional"`
	Format string `query:"optional"`
}

// OpSave saves an object to local storage and returns its ID. The caller must provide the exact size
// of the payload and send exactly that many bytes after the query is accepted. An error or an objectID
// will be sent back. If Type is not empty, a standard object header will be written before the payload.
func (mod *Module) OpSave(ctx *astral.Context, q shell.Query, args opSaveArgs) (err error) {
	if args.Size < 0 || int64(args.Size) > MaxObjectSize {
		return q.Reject()
	}

	ch := astral.NewChannel(q.Accept(), args.Format)
	defer ch.Close()
	
	var payload = make([]byte, args.Size)
	_, err = io.ReadFull(ch.Transport(), payload)
	if err != nil {
		mod.log.Errorv(1, "op_save: error reading object: %v", err)
		return nil
	}

	raw := &astral.RawObject{
		Type:    args.Type,
		Payload: payload,
	}

	objID, err := mod.Save(ctx.WithIdentity(q.Caller()), raw)
	if err != nil {
		ch.Write(astral.NewError("internal error"))
		mod.log.Errorv(1, "op_save: error writing object: %v", err)
		return err
	}

	return ch.Write(objID)
}
