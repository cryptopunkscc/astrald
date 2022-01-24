package service

import astral "github.com/cryptopunkscc/astrald/mod/apphost/client"

func (srv *Context) IsRejected(request astral.Request) bool {
	caller := request.Caller()
	isRemote := caller != "" && caller != srv.Identity
	if isRemote {
		request.Reject()
		srv.Println("Accept only local request, but caller was remote", caller)
	}
	return isRemote
}
