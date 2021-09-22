package tor

type Service struct {
	ctl        *Control
	serviceID  string
	privateKey string
}

func (s *Service) PrivateKey() string {
	return s.privateKey
}

func (s *Service) ServiceID() string {
	return s.serviceID
}

func (s *Service) Close() error {
	return s.ctl.stopService(s.serviceID)
}
