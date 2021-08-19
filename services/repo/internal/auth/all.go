package auth

import "github.com/cryptopunkscc/astrald/api"

func AcceptAll(_ api.Core, _ api.ConnectionRequest) bool {
	return true
}
