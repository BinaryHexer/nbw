package stream

import (
	"io"
	"log"

	"github.com/reugn/go-streams"
	ext "github.com/reugn/go-streams/extension"

	"github.com/BinaryHexer/nbw/pkg/stream"
	"sync"
)

var bufPool = &sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 500)
	},
}

type Writer struct {
	w    io.Writer
	in   chan interface{}
	out  chan interface{}
	done chan struct{}
}

func NewWriter(w io.Writer, flows []streams.Flow) *Writer {
	wr := &Writer{
		w:    w,
		in:   make(chan interface{}),
		out:  make(chan interface{}),
		done: make(chan struct{}),
	}

	go wr.write()
	go wr.init(flows)

	return wr
}

func (w *Writer) Write(p []byte) (int, error) {
	// p might be pooled so we avoid to hold a reference to it, and instead copy it.
	p = append(bufPool.Get().([]byte), p...)
	w.in <- p

	return len(p), nil
}

func (w *Writer) Close() error {
	// close the input channel
	close(w.in)

	// wait processing
	<-w.done

	// close if the underlying writer supports it
	if w, ok := w.w.(io.Closer); ok {
		return w.Close()
	}

	return nil
}

func (w *Writer) init(flows []streams.Flow) {
	source := ext.NewChanSource(w.in)
	sink := ext.NewChanSink(w.out)
	f := stream.NewPipe(flows...)

	source.
		Via(f).
		To(sink)
}

func (w *Writer) write() {
	for e := range w.out {
		p := e.([]byte)
		_, err := w.w.Write(p)
		if err != nil {
			log.Printf("err occurred: %v\n", err)
		}

		// Proper usage of a sync.Pool requires each entry to have approximately
		// the same memory cost. To obtain this property when the stored type
		// contains a variably-sized buffer, we add a hard limit on the maximum buffer
		// to place back in the pool.
		//
		// See https://golang.org/issue/23199
		const maxSize = 1 << 16 // 64KiB
		if cap(p) <= maxSize {
			bufPool.Put(p[:0])
		}
	}
	close(w.done)
}
