package io

import (
	"io"
)

// A WriteSyncer is an io.Writer that can also flush any buffered data.
type WriteSyncer interface {
	io.Writer

	Sync() error
}

// A WriteCloserSync is an io.Writer that implements both io.WriteCloser and Sync.
type WriteCloserSync interface {
	io.Writer

	io.WriteCloser
	WriteSyncer
}

func AddCloserSync(w io.Writer) WriteCloserSync {
	switch w := w.(type) {
	case io.WriteCloser:
		return writeCloser{w}
	case WriteSyncer:
		return writeSyncer{w}
	default:
		return writerWrapper{w}
	}
}

type writeCloser struct {
	io.WriteCloser
}

func (w writeCloser) Sync() error {
	return w.Close()
}

type writeSyncer struct {
	WriteSyncer
}

func (w writeSyncer) Close() error {
	return w.Sync()
}

type writerWrapper struct {
	io.Writer
}

func (w writerWrapper) Close() error {
	return nil
}

func (w writerWrapper) Sync() error {
	return nil
}
