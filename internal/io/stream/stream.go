package stream

import (
	"io"
	"log"

	"github.com/reugn/go-streams"
	ext "github.com/reugn/go-streams/extension"

	"github.com/BinaryHexer/nbw/pkg/stream"
)

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
		b := e.([]byte)
		_, err := w.w.Write(b)
		if err != nil {
			log.Printf("err occurred: %v\n", err)
		}
	}
	close(w.done)
}
