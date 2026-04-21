package nodes

// ErrBufferEmpty is returned by InputBuffer.Read and OutputBuffer.Write when no progress
// can be made. Wait on ErrBufferEmpty.Wait() and retry.
type ErrBufferEmpty struct {
	ch <-chan struct{}
}

func (e *ErrBufferEmpty) Error() string         { return "buffer empty" }
func (e *ErrBufferEmpty) Wait() <-chan struct{} { return e.ch }
