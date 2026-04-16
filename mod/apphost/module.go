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
	MethodCancel          = "apphost.cancel"
	MethodBind            = "apphost.bind"
	MethodNewAppContract  = "apphost.new_app_contract"
	MethodSignAppContract = "apphost.sign_app_contract"
	MethodInstallApp      = "apphost.install_app"
)

type Module interface {
	CreateAccessToken(*astral.Identity, astral.Duration) (*AccessToken, error)
	LocalApps() ([]*App, error)
}

var ErrProtocolError = errors.New("protocol error")
