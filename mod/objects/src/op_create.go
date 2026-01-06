package objects

import (
	"errors"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opCreateArgs struct {
	Alloc uint64 `query:"optional"`
	Repo  string `query:"optional"`
	In    string `query:"optional"`
	Out   string `query:"optional"`
}

// OpCreate creates a new object in the repository. It expects a stream of Blob objects followed by objects.Commit.
// On successful commit returns an ObjectID, an ErrorMessage otherwise. Closing the connection before committing
// will discard the data.
func (mod *Module) OpCreate(ctx *astral.Context, q shell.Query, args opCreateArgs) (err error) {
	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	var repo = mod.WriteDefault()

	// check repo
	if args.Repo != "" {
		repo = mod.GetRepository(args.Repo)
		if repo == nil {
			return ch.Send(astral.NewError("repository not found"))
		}
	}

	// create a new object in the repo
	w, err := repo.Create(ctx, &objects.CreateOpts{Alloc: int(args.Alloc)})
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}
	defer w.Discard() // make sure we don't leave garbage behind

	// send an ack
	err = ch.Send(&astral.Ack{})
	if err != nil {
		return
	}

	return ch.Collect(func(msg astral.Object) (err error) {
		switch msg := msg.(type) {
		case *astral.Blob:
			_, err = msg.WriteTo(w)

		case *objects.CommitMsg: // commit the object
			objectID, err := w.Commit()

			if err != nil {
				return ch.Send(astral.NewError(err.Error()))
			}

			mod.log.Logv(3, "%v created %v in %v", q.Caller(), objectID, repo)

			return ch.Send(objectID)

		default:
			return errors.New("unexpected message type")
		}
		return
	})
}
