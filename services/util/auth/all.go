package auth

import "github.com/cryptopunkscc/astrald/api"

func All(_ api.Core, _ api.ConnectionRequest) bool {
	return true
}
