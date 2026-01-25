package keys

import (
	"encoding/hex"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
)

type opSignASN1Args struct {
	Hash string
	Out  string `query:"optional"`
}

// Obsolete: this op needs to be rewritten
func (mod *Module) OpSignASN1(_ *astral.Context, q *ops.Query, args opSignASN1Args) (err error) {
	hash, err := hex.DecodeString(args.Hash)
	if err != nil {
		return q.RejectWithCode(astral.CodeInvalidQuery)
	}
	signer := q.Caller()

	sig, err := mod.SignASN1(signer, hash)
	if err != nil {
		return q.Reject()
	}

	ch := channel.New(q.Accept(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	return ch.Send((*astral.Bytes8)(&sig))
}
