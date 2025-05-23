package status

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

func (ops *Ops) Show(ctx *astral.Context, q shell.Query) (err error) {
	if v, _ := q.Extra().Get("interface"); v != "terminal" {
		return q.Reject()
	}

	t := shell.NewTerminal(q.Accept())
	defer t.Close()

	for k, v := range ops.mod.Cache().Clone() {
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
