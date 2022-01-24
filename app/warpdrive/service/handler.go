package service

import astral "github.com/cryptopunkscc/astrald/mod/apphost/client"

type Handler func(srv Context, request astral.Request)

type Handlers map[string]Handler
