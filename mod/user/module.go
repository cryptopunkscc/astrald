package user

import (
	"github.com/cryptopunkscc/astrald/astral"
)

const ModuleName = "user"
const DBPrefix = "users__"

const (
	OpSyncAssets      = "user.sync_assets"
	OpInfo            = "user.info"
	OpAddAsset        = "user.add_asset"
	OpRemoveAsset     = "user.remove_asset"
	OpInvite          = "user.invite"
	OpClaim           = "user.claim"
	OpRequestInvite   = "user.request_invite"
	OpSwarmStatus     = "user.swarm_status"
	OpCreate          = "user.create"
	OpListSiblings    = "user.list_siblings"
	OpAssets          = "user.assets"
	OpNewNodeContract = "user.new_node_contract"
	OpSyncWith        = "user.sync_with"
)

type Module interface {
	Ready() <-chan struct{}
	Identity() *astral.Identity
	LocalSwarm() (list []*astral.Identity)
	NewMaintainLinkTask(target *astral.Identity) MaintainLinkTask
	NewSyncNodesTask(remoteIdentity *astral.Identity) SyncNodesAction
}
