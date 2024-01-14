package keys

import "context"

type Service struct {
	*Module
}

func (srv *Service) Run(ctx context.Context) error {
	return nil
}
