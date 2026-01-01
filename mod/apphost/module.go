package apphost

import (
	"errors"

	"github.com/cryptopunkscc/astrald/astral"
)

const ModuleName = "apphost"
const DBPrefix = "apphost__"

type Module interface {
	CreateAccessToken(*astral.Identity, astral.Duration) (*AccessToken, error)
	ActiveLocalAppContracts() ([]*AppContract, error)
}

var ErrProtocolError = errors.New("protocol error")
