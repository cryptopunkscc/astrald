package apphost

import (
	"errors"

	"github.com/cryptopunkscc/astrald/astral"
)

const ModuleName = "apphost"
const DBPrefix = "apphost__"

const (
	MethodCreateToken     = "apphost.create_token"
	MethodListTokens      = "apphost.list_tokens"
	MethodRegisterHandler = "apphost.register_handler"
	MethodNewAppContract  = "apphost.new_app_contract"
	MethodSignAppContract = "apphost.sign_app_contract"
	MethodIndex           = "apphost.index"
	MethodCancel          = "apphost.cancel"
	MethodBind            = "apphost.bind"
)

type Module interface {
	CreateAccessToken(*astral.Identity, astral.Duration) (*AccessToken, error)
	ActiveLocalAppContracts() ([]*SignedAppContract, error)
}

var ErrProtocolError = errors.New("protocol error")
var ErrInactiveContract = errors.New("inactive contract")
