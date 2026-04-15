package objects

import (
	"reflect"
	"slices"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	libquery "github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/lib/routing"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

type opSpecArgs struct {
	Type string `query:"optional"`
	In   string `query:"optional"`
	Out  string `query:"optional"`
}

func (mod *Module) OpSpec(ctx *astral.Context, query *routing.IncomingQuery, args opSpecArgs) (err error) {
	ch := query.Accept(channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	types := astral.DefaultBlueprints().Types()
	slices.Sort(types)
	for _, typeName := range types {
		if len(args.Type) > 0 && typeName != args.Type {
			continue
		}

		obj := astral.New(typeName)
		if obj == nil {
			continue
		}

		v := reflect.ValueOf(obj)
		if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
			continue
		}

		editor := libquery.EditCamel(obj)

		spec := &objects.TypeSpec{
			Name:   typeName,
			Fields: editor.Spec(),
		}

		if len(spec.Fields) == 0 {
			continue
		}

		err = ch.Send(spec)
		if err != nil {
			return err
		}
	}

	return ch.Send(&astral.EOS{})
}
