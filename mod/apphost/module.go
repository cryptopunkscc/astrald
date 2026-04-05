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
	MethodSignAppContract = "apphost.sign_app_contract"
	MethodCancel          = "apphost.cancel"
	MethodBind            = "apphost.bind"
)

type Module interface {
	CreateAccessToken(*astral.Identity, astral.Duration) (*AccessToken, error)
	ActiveLocalAppContracts() ([]*AppContract, error)
}

var ErrProtocolError = errors.New("protocol error")
var ErrNodeUnavailable = errors.New("node unavailable")
