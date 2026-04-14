package apphost

import (
	"errors"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/auth"
)

const ModuleName = "apphost"
const DBPrefix = "apphost__"

const (
	MethodCreateToken     = "apphost.create_token"
	MethodListTokens      = "apphost.list_tokens"
	MethodRegisterHandler = "apphost.register_handler"
	MethodSignAppContract = "apphost.sign_app_contract"
	MethodCancel          = "apphost.cancel"
	MethodBind            = "apphost.bind"
	MethodIndex           = "apphost.index"
	MethodNewAppContract  = "apphost.new_app_contract"
)

type Module interface {
	CreateAccessToken(*astral.Identity, astral.Duration) (*AccessToken, error)
	ActiveLocalAppContracts() ([]*auth.SignedContract, error)
}

var ErrProtocolError = errors.New("protocol error")
