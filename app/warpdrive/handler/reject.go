package handler

import astral "github.com/cryptopunkscc/astrald/mod/apphost/client"

func (ctx *Context) IsRejected(request astral.Request) bool {
	caller := request.Caller()
	isRemote := caller != "" && caller != ctx.Identity
	if isRemote {
		request.Reject()
		ctx.Println("Accept only local request, but caller was remote", caller)
	}
	return isRemote
}
