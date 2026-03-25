package gateway

import "github.com/cryptopunkscc/astrald/astral"

type Visibility = astral.String8

const (
	VisibilityPublic  Visibility = "public"
	VisibilityPrivate Visibility = "private"
)
