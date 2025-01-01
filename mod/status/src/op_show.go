package status

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

func (mod *Module) opShow(ctx astral.Context, q shell.Query) (err error) {
	t, err := shell.AcceptTerminal(q)
	if err != nil {
		return err
	}
	defer t.Close()

	for k, v := range mod.Cache().Clone() {
		attachments := v.Status.Attachments.Objects()

		t.Printf("%v:%v - %v (%v), %v objects\n",
			k,
			uint16(v.Status.Port),
			v.Status.Alias,
			v.Identity,
			len(attachments),
		)

		for _, a := range attachments {
			id, err := astral.ResolveObjectID(a)
			if err != nil {
				return err
			}
			t.Printf("- %v (%v)\n", a.ObjectType(), id)
		}
	}

	return
}
