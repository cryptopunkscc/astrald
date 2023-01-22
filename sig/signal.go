package sig

var _ Signal = Sig(nil)

type Sig <-chan struct{}

func New() chan struct{} {
	return make(chan struct{}, 1)
}

func (s Sig) Done() <-chan struct{} {
	return s
}
