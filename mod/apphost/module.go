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
	MethodHoldObject      = "apphost.hold_object"
	MethodUnholdObject    = "apphost.unhold_object"
	MethodListHeldObjects = "apphost.list_held_objects"
)

// Module is the public API surface of the apphost module.
type Module interface {
	CreateAccessToken(*astral.Identity, astral.Duration) (*AccessToken, error)
	LocalApps() ([]*App, error)
}

var ErrProtocolError = errors.New("protocol error")
var ErrMissingAppIdentity = errors.New("missing app identity")
var ErrMissingObjectID = errors.New("missing object id")
