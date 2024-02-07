package presence

const ModuleName = "presence"

type Module interface {
	SetVisible(bool) error
}

const (
	DiscoverFlag = "discover"
	PairingFlag  = "pair"
)
