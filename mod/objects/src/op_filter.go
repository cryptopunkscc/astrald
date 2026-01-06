package objects

import (
	"errors"
	"io"
	"slices"
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opFilterArgs struct {
	Only   *string `query:"optional"`
	Except *string `query:"optional"`
	In     string  `query:"optional"`
	Out    string  `query:"optional"`
}

func (mod *Module) OpFilter(ctx *astral.Context, q shell.Query, args opFilterArgs) (err error) {
	// prepare lists
	var only, except []string
	if args.Only != nil && len(*args.Only) > 0 {
		only = strings.Split(*args.Only, ",")
	}
	if args.Except != nil && len(*args.Except) > 0 {
		except = strings.Split(*args.Except, ",")
	}

	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out), channel.AllowUnparsed(true))
	defer ch.Close()

	for {
		object, err := ch.Receive()
		switch {
		case err == nil:
		case errors.Is(err, io.EOF):
			return nil
		case errors.Is(err, astral.ErrBlueprintNotFound{}):
			continue
		default:
			return err
		}

		if len(only) > 0 {
			if !slices.Contains(only, object.ObjectType()) {
				continue
			}
		}

		if len(except) > 0 {
			if slices.Contains(only, object.ObjectType()) {
				continue
			}
		}

		err = ch.Send(object)
		if err != nil {
			return err
		}
	}
}
