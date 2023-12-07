package streams

import "io"

// Skip skips n bytes on the reader
func Skip(reader io.Reader, n uint64) error {
	var buf [4096]byte
	var left = n

	for left > 0 {
		var chunkSize = min(4096, left)

		nn, err := reader.Read(buf[:chunkSize])
		if err != nil {
			return err
		}

		left -= uint64(nn)
	}

	return nil
}
