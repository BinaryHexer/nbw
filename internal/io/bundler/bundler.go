package bundler

import (
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	bbi "github.com/BinaryHexer/nbw/internal/bundler"
	bbx "github.com/BinaryHexer/nbw/pkg/bundler"
	"google.golang.org/api/support/bundler"
)

const (
	DefaultErrChanCapacity = 10
	errWriteErr            = "failed to write: %v"
)

//nolint:gochecknoglobals  // necessary to maintain bufPool(byte) pool and errors
var (
	bufPool = &sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, 500)
		},
	}
	errClosed = errors.New("writer already closed")
)

// WriterOption can be used to setup the writer.
type WriterOption func(*Writer) bbx.Option

// WithDelayThreshold sets the interval at which the bundler is flushed.
// The default is DefaultDelayThreshold.
func WithDelayThreshold(delay time.Duration) WriterOption {
	return WriterOption(func(bw *Writer) bbx.Option {
		return bbx.WithDelayThreshold(delay)
	})
}

// WithBundleCountThreshold sets the max number of items after which the bundler is flushed.
// The default is DefaultBundleCountThreshold.
func WithBundleCountThreshold(n int) WriterOption {
	return WriterOption(func(bw *Writer) bbx.Option {
		return bbx.WithBundleCountThreshold(n)
	})
}

// WithBundleByteThreshold sets the max size of the bundle (in bytes) at which the
// bundler is flushed. The default is DefaultBundleByteThreshold.
func WithBundleByteThreshold(n int) WriterOption {
	return WriterOption(func(bw *Writer) bbx.Option {
		return bbx.WithBundleByteThreshold(n)
	})
}

// WithBundleByteLimit sets the maximum size of a bundle, in bytes.
// Zero means unlimited. The default is DefaultBundleByteLimit.
func WithBundleByteLimit(n int) WriterOption {
	return WriterOption(func(bw *Writer) bbx.Option {
		return bbx.WithBundleByteLimit(n)
	})
}

// WithBufferedByteLimit sets the maximum number of bytes that the Bundler will keep
// in memory before returning ErrOverflow. The default is DefaultBufferedByteLimit.
func WithBufferedByteLimit(n int) WriterOption {
	return WriterOption(func(bw *Writer) bbx.Option {
		return bbx.WithBufferedByteLimit(n)
	})
}

// WithOnError sets the function to be executed on errors.
// The default is a simple log.
func WithOnError(f func(err error)) WriterOption {
	return WriterOption(func(bw *Writer) bbx.Option {
		bw.onError = f

		return nil
	})
}

// WithErrorChannelCapacity sets the buffer capacity of errors channel.
// The default is DefaultErrChanCapacity.
func WithErrorChannelCapacity(n int) WriterOption {
	return WriterOption(func(bw *Writer) bbx.Option {
		bw.errs = make(chan error, n)

		return nil
	})
}

// Writer is a io.Writer wrapper that uses a bundler to make Write lock-free,
// non-blocking and thread safe.
type Writer struct {
	w       io.Writer
	b       *bundler.Bundler
	errs    chan error
	onError func(err error)
}

// NewWriter creates a writer wrapping w with a bundler in order to never block
// the producers and drop writes if the underlying writer can't keep up with the
// flow of data.
//
// Use a bundler.Writer when
//
// 	   onError := WithOnError(func(err error) {
// 	   	   log.Printf("Dropped writes due to: %v", err)
// 	   })
// 	   wr := NewWriter(w, []WriterOption{onError})
// 	   wr.Write([]byte("Hello, World!"))
//
// See https://pkg.go.dev/google.golang.org/api/support/bundler for more info on bundler.
func NewWriter(w io.Writer, opts []WriterOption) *Writer {
	bw := &Writer{
		w:    w,
		errs: make(chan error, DefaultErrChanCapacity),
		onError: func(err error) {
			log.Printf("Dropped writes due to: %v", err)
		},
	}

	bOpts := bw.applyOpts(opts)
	b := bw.newBundler(bOpts...)
	bw.b = b

	return bw
}

func (bw *Writer) applyOpts(opts []WriterOption) []bbx.Option {
	bbOpts := make([]bbx.Option, 0)

	for _, o := range opts {
		bbOpt := o(bw)
		if bbOpt != nil {
			bbOpts = append(bbOpts, bbOpt)
		}
	}

	return bbOpts
}

func (bw *Writer) newBundler(opts ...bbx.Option) *bundler.Bundler {
	b := bbi.NewBundler(&[]byte{}, func(p interface{}) {
		xs := p.([]*[]byte)
		for _, x := range xs {
			b := *x
			bw.write(b)
		}
	})

	for _, o := range opts {
		o(b)
	}

	return b
}

func (bw *Writer) Write(p []byte) (int, error) {
	// when the writer is closed, b will be nil.
	if bw.b == nil {
		return 0, errClosed
	}

	// copy slice here because byte buffer may changes before bundler flush the byte slice
	// more memory allocations but it should be fast because write operation is non-blocking and slice copy is 60 ns/op operation
	q := append(bufPool.Get().([]byte), p...)

	// write to the bundler
	if err := bw.b.Add(&q, len(q)); err != nil {
		bw.error(fmt.Errorf(errWriteErr, err))
	}

	return len(q), nil
}

func (bw *Writer) Close() error {
	bb := bw.b

	// unset the bundler so no further writes can occur
	bw.b = nil

	// flush the bundler
	bb.Flush()

	// close if the underlying writer supports it
	if w, ok := bw.w.(io.Closer); ok {
		return w.Close()
	}

	return nil
}

func (bw *Writer) write(p []byte) {
	_, err := bw.w.Write(p)
	if err != nil {
		bw.error(fmt.Errorf(errWriteErr, err))
	}

	// Proper usage of a sync.Pool requires each entry to have approximately
	// the same memory cost. To obtain this property when the stored type
	// contains a variably-sized buffer, we add a hard limit on the maximum buffer
	// to place back in the pool.
	//
	// See https://golang.org/issue/23199
	const maxSize = 1 << 16 // 64KiB
	if cap(p) <= maxSize {
		// reset the byte slice
		bufPool.Put(p[:0])
	}
}

func (bw *Writer) error(err error) {
	bw.errs <- err
}

func (bw *Writer) handleErrors(err error) {
	for {
		select {
		case bw.errs <- err:
			bw.onError(err)
		}
	}
}
