package ip

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opPublicIPCandidatesArgs struct {
	Out string `query:"optional"`
}

func (mod *Module) OpPublicIPCandidates(ctx *astral.Context, q shell.Query, args opPublicIPCandidatesArgs) (err error) {
	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	for _, addr := range mod.PublicIPCandidates() {
		err = ch.Write(&addr)
		if err != nil {
			return
		}
	}
	return
}
