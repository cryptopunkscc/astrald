package streams

type NilWriter struct {
}

func (NilWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

type NilCloser struct {
}

func (NilCloser) Close() error {
	return nil
}

type NilWriteCloser struct {
	NilWriter
	NilCloser
}
