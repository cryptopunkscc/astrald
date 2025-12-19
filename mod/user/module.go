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
	ActionRevokeContract = "user.revoke_contract"
)

type Module interface {
	LocalSwarm() (list []*astral.Identity)
	NewMaintainLinkAction(target *astral.Identity) MaintainLinkAction
	NewSyncNodesAction(remoteIdentity *astral.Identity) SyncNodesAction
}
