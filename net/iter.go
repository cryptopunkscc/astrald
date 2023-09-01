package net

import "io"

type WriterIter interface {
	NextWriter() io.Writer
}

type ReaderIter interface {
	NextReader() io.Reader
}

func FinalWriter(w io.Writer) (final io.Writer) {
	final = w
	for {
		if n, ok := final.(WriterIter); ok {
			next := n.NextWriter()
			if next == nil {
				return
			}
			final = next
		} else {
			return
		}
	}
}

func FinalReader(w io.Reader) (final io.Reader) {
	final = w
	for {
		if n, ok := final.(ReaderIter); ok {
			next := n.NextReader()
			if next == nil {
				return
			}
			final = next
		} else {
			return
		}
	}
}
