package user

import (
	"encoding/json"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"time"
)

type opCreateArgs struct {
	Alias string
	Force bool `query:"optional"`
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

	conn, err := q.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()

	// create a private key for the user
	userID, keyID, err := mod.Keys.CreateKey(args.Alias)
	enc := json.NewEncoder(conn)
	enc.SetIndent("", "  ")

	if err != nil {
		return enc.Encode(map[string]string{
			"error": fmt.Sprintf("create user key: %s", err.Error()),
		})
	}

	// create an access token
	token, err := mod.Apphost.CreateAccessToken(userID, astral.Duration(100*365*24*time.Hour))
	if err != nil {
		return enc.Encode(map[string]string{
			"error": fmt.Sprintf("create access token: %s", err.Error()),
		})
	}

	// sign a node contrdct with the user
	contract, err := mod.SignLocalContract(userID)
	if err != nil {
		return enc.Encode(map[string]string{
			"error": fmt.Sprintf("sign contract: %s", err.Error()),
		})
	}

	// set the contract as active
	err = mod.SetActiveContract(contract)
	if err != nil {
		if err != nil {
			return enc.Encode(map[string]string{
				"error": fmt.Sprintf("set active contract: %s", err.Error()),
			})
		}
	}

	// return a summary
	contractID, _ := astral.ResolveObjectID(contract)

	return enc.Encode(map[string]interface{}{
		"user_alias":   args.Alias,
		"user_id":      userID,
		"key_id":       keyID,
		"contract_id":  contractID,
		"contract":     contract,
		"access_token": token.Token,
	})
}
