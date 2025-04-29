package keys

import (
	"encoding/hex"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opSignASN1Args struct {
	Hash string
	Out  string `query:"optional"`
}

func (mod *Module) OpSignASN1(_ *astral.Context, q shell.Query, args opSignASN1Args) (err error) {
	hash, err := hex.DecodeString(args.Hash)
	if err != nil {
		return q.RejectWithCode(astral.CodeInvalidQuery)
	}
	signer := q.Caller()

	sig, err := mod.SignASN1(signer, hash)
	if err != nil {
		return q.Reject()
	}

	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	return ch.WritePayload((*astral.Bytes8)(&sig))
}
