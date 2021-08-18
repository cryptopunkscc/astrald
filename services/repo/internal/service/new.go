package service

import "github.com/cryptopunkscc/astrald/api"

func New(
	port string,
	authorize Authorize,
	handlers Handlers,
) *Service {
	return &Service{
		port:      port,
		observers: map[api.Stream]struct{}{},
		authorize: authorize,
		handlers:  handlers,
	}
}