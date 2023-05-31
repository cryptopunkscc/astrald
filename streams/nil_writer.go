package streams

type NilWriter struct {
}

func (NilWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}
