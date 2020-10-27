package io

import (
	"io"
)

func AddCloser(w io.Writer) io.WriteCloser {
	switch w := w.(type) {
	case io.WriteCloser:
		return w
	default:
		return writerWrapper{w}
	}
}

type writerWrapper struct {
	io.Writer
}

func (w writerWrapper) Close() error {
	return nil
}
