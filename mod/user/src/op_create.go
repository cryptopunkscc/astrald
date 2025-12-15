package user

import (
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/mod/user"
)

type opCreateArgs struct {
	Alias string
	Force bool `query:"optional"`

	In  string `query:"optional"`
	Out string `query:"optional"`
}

// OpCreate creates a new user with provided alias, signs a node contract between the new user and the local node and
// sets that contract as active. It rejects if there's an active contract unless force is true.
func (mod *Module) OpCreate(ctx *astral.Context, q shell.Query, args opCreateArgs) (err error) {
	// reject network calls
	if q.Origin() == "network" {
		return q.Reject()
	}

	// only allow this if there's no currently active contract
	if mod.ActiveContract() != nil && !args.Force {
		return q.Reject()
	}

	ch := astral.NewChannelFmt(q.Accept(), args.In, args.Out)
	defer ch.Close()

	// create a private key for the user
	userID, keyID, err := mod.Keys.CreateKey(args.Alias)
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	// create an access token
	token, err := mod.Apphost.CreateAccessToken(userID, astral.Duration(100*365*24*time.Hour))
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	// sign a node contrdct with the user
	contract, err := mod.SignLocalContract(userID)
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	// set the contract as active
	err = mod.SetActiveContract(contract)
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	// return a summary
	contractID, _ := astral.ResolveObjectID(contract)
	userInfo := user.CreatedUserInfo{
		ID:          userID,
		Alias:       astral.String8(args.Alias),
		KeyID:       keyID,
		ContractID:  contractID,
		AccessToken: token.Token,
		Contract:    contract,
	}

	return ch.Write(&userInfo)
}
