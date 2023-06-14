package main

// this file includes all modules that should be compiled into the node

import (
	_ "github.com/cryptopunkscc/astrald/mod/admin"
	_ "github.com/cryptopunkscc/astrald/mod/agent"
	_ "github.com/cryptopunkscc/astrald/mod/apphost"
	_ "github.com/cryptopunkscc/astrald/mod/connect"
	_ "github.com/cryptopunkscc/astrald/mod/discovery"
	_ "github.com/cryptopunkscc/astrald/mod/gateway"
	_ "github.com/cryptopunkscc/astrald/mod/keepalive"
	_ "github.com/cryptopunkscc/astrald/mod/optimizer"
	_ "github.com/cryptopunkscc/astrald/mod/presence"
	_ "github.com/cryptopunkscc/astrald/mod/profile"
	_ "github.com/cryptopunkscc/astrald/mod/reflectlink"
	//_ "github.com/cryptopunkscc/astrald/mod/roam"
	_ "github.com/cryptopunkscc/astrald/mod/shift"
	_ "github.com/cryptopunkscc/astrald/mod/storage"
	_ "github.com/cryptopunkscc/astrald/mod/tcpfwd"
)
