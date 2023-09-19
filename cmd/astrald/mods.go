package main

// this file includes all modules that should be compiled into the node

import (
	_ "github.com/cryptopunkscc/astrald/mod/admin"
	_ "github.com/cryptopunkscc/astrald/mod/agent"
	_ "github.com/cryptopunkscc/astrald/mod/apphost"
	_ "github.com/cryptopunkscc/astrald/mod/connect"
	_ "github.com/cryptopunkscc/astrald/mod/fwd"
	_ "github.com/cryptopunkscc/astrald/mod/gateway"
	_ "github.com/cryptopunkscc/astrald/mod/keepalive"
	_ "github.com/cryptopunkscc/astrald/mod/presence"
	_ "github.com/cryptopunkscc/astrald/mod/profile"
	_ "github.com/cryptopunkscc/astrald/mod/reflectlink"
	_ "github.com/cryptopunkscc/astrald/mod/router"
	_ "github.com/cryptopunkscc/astrald/mod/sdp"
	_ "github.com/cryptopunkscc/astrald/mod/speedtest"
	_ "github.com/cryptopunkscc/astrald/mod/storage"
)
