package user

import (
	"github.com/cryptopunkscc/astrald/astral"
)

const ModuleName = "user"
const DBPrefix = "users__"

const (
	OpSyncAssets  = "user.sync_assets"
	OpInfo        = "user.info"
	OpAddAsset    = "user.add_asset"
	OpRemoveAsset = "user.remove_asset"
	OpInvite      = "user.invite"
	OpLink        = "user.link"
	OpClaim       = "user.claim"
)

type Module interface {
	NewMaintainLinkAction(target *astral.Identity) MaintainLinkAction
	NewEnsureConnectivityAction() EnsureConnectivityAction
}
