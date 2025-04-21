package user

const ModuleName = "user"
const DBPrefix = "users__"

const (
	OpSyncAssets  = "user.sync_assets"
	OpAddAsset    = "user.add_asset"
	OpRemoveAsset = "user.remove_asset"
	OpInvite      = "user.invite"
	OpLink        = "user.link"
	OpClaim       = "user.claim"
)

type Module interface {
}
