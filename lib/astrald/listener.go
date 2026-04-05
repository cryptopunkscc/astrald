package astrald

import "net"

// HandlerListener adapts a Handler to the net.Listener interface.
type HandlerListener struct {
	*Handler
}

var _ net.Listener = &HandlerListener{}

func NewHandlerListener(h *Handler) *HandlerListener {
	return &HandlerListener{Handler: h}
}

func (l *HandlerListener) Accept() (net.Conn, error) {
	pending, err := l.Handler.ReadQuery()
	if err != nil {
		return nil, err
	}
	return pending.Accept(), nil
}

func (h *Handler) Addr() net.Addr {
	return h.listener.Addr()
}
