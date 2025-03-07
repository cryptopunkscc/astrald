package keys

import (
	"encoding/json"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opCreateKeyArgs struct {
	Alias  astral.String
	Format astral.String
}

func (mod *Module) OpCreateKey(_ astral.Context, q shell.Query, args opCreateKeyArgs) (err error) {
	if args.Alias == "" {
		return q.Reject()
	}

	mod.log.Infov(1, "creating key for %v", args.Alias)

	key, _, err := mod.CreateKey(args.Alias.String())
	if err != nil {
		mod.log.Errorv(1, "error creating key for %v: %v", args.Alias, err)
		return q.Reject()
	}

	conn, err := q.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()

	switch args.Format {
	case "json":
		err = json.NewEncoder(conn).Encode(key)
	case "bin", "":
		_, err = key.WriteTo(conn)
	default:
		return errors.New("unsupported format")
	}
	return
}
