package user

import (
	"github.com/cryptopunkscc/astrald/astral"
)

const ModuleName = "user"
const DBPrefix = "users__"

const (
	OpSyncAssets    = "user.sync_assets"
	OpInfo          = "user.info"
	OpAddAsset      = "user.add_asset"
	OpRemoveAsset   = "user.remove_asset"
	OpInvite        = "user.invite"
	OpClaim         = "user.claim"
	OpRequestInvite = "user.request_invite"
	OpSwarmStatus   = "user.swarm_status"
	OpCreate        = "user.create"
	OpListSiblings  = "user.list_siblings"
	OpAssets        = "user.assets"
	OpAddToIndex    = "user.add_to_index"
	OpNewContract   = "user.new_contract"
	OpSignContract  = "user.sign_contract"
	OpSyncWith      = "user.sync_with"

	ActionSwarmAccess = "mod.user.swarm_access_action" // equals SwarmAccessAction{}.ObjectType()
)

type Module interface {
	Identity() *astral.Identity
	LocalSwarm() (list []*astral.Identity)
	NewMaintainLinkTask(target *astral.Identity) MaintainLinkTask
	NewSyncNodesTask(remoteIdentity *astral.Identity) SyncNodesAction
}
