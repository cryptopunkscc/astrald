package user

import (
	"github.com/cryptopunkscc/astrald/astral"
)

const ModuleName = "user"
const DBPrefix = "users__"

const (
	OpSyncAssets         = "user.sync_assets"
	OpInfo               = "user.info"
	OpAddAsset           = "user.add_asset"
	OpRemoveAsset        = "user.remove_asset"
	OpInvite             = "user.invite"
	OpLink               = "user.link"
	OpClaim              = "user.claim"
	OpRequestInvite      = "user.request_invite"
	OpRevokeNodeContract = "user.revoke_node_contract"
	OpSwarmStatus        = "user.swarm_status"
	OpCreate             = "user.create"
	OpListSiblings       = "user.list_siblings"
	OpAssets             = "user.assets"
	OpAddToIndex         = "user.add_to_index"
	OpNewNodeContract    = "user.new_node_contract"
	OpSignNodeContract   = "user.sign_node_contract"
	OpSyncWith           = "user.sync_with"
	ActionRevokeContract = "user.revoke_contract"
)

type Module interface {
	LocalSwarm() (list []*astral.Identity)
	NewMaintainLinkTask(target *astral.Identity) MaintainLinkTask
	NewSyncNodesTask(remoteIdentity *astral.Identity) SyncNodesAction
}
