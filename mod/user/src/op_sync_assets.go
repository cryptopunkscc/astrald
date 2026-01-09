package user

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opSyncAssetsArgs struct {
	Start int    `query:"optional"`
	Out   string `query:"optional"`
}

func (mod *Module) OpSyncAssets(ctx *astral.Context, q shell.Query, args opSyncAssetsArgs) (err error) {
	var rows []*dbAsset

	err = mod.db.Where("height >= ?", args.Start).Find(&rows).Error
	if err != nil {
		mod.log.Error("db error: %v", err)
		return q.RejectWithCode(2)
	}

	ch := channel.New(q.Accept(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	var height astral.Uint64

	if len(rows) == 0 {
		height = astral.Uint64(args.Start)
	} else {
		for _, row := range rows {
			height = max(height, astral.Uint64(row.Height))

			err = ch.Send(&OpUpdate{
				Nonce:    row.Nonce,
				ObjectID: row.ObjectID,
				Removed:  astral.Bool(row.Removed),
			})
		}
		height++
	}

	return ch.Send(&height)
}

type OpUpdate struct {
	Nonce    astral.Nonce
	ObjectID *astral.ObjectID
	Removed  astral.Bool
}

var _ astral.Object = &OpUpdate{}

func (s OpUpdate) ObjectType() string { return "mod.user.op_update" }

func (s OpUpdate) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(s).WriteTo(w)
}

func (s *OpUpdate) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(s).ReadFrom(r)
}

func init() {
	_ = astral.Add(&OpUpdate{})
}
