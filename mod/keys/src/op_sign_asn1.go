package keys

import (
	"encoding/hex"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opSignASN1Args struct {
	Hash   astral.String
	Format astral.String `query:"optional"`
}

func (mod *Module) OpSignASN1(_ astral.Context, q shell.Query, args opSignASN1Args) (err error) {
	hash, err := hex.DecodeString(args.Hash.String())
	if err != nil {
		return q.Reject()
	}
	signer := q.Caller()

	sig, err := mod.SignASN1(signer, hash)
	if err != nil {
		return q.Reject()
	}

	c, _ := q.Accept()
	defer c.Close()

	switch args.Format {
	case "json":
		json.NewEncoder(c).Encode(sig)
	case "bin", "":
		c.Write(sig)
	}

	return nil
}
