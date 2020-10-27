package nbw

import (
	"io"
	"time"

	"github.com/reugn/go-streams"

	"github.com/BinaryHexer/nbw/internal/io/bundler"
	"github.com/BinaryHexer/nbw/internal/io/diode"
	"github.com/BinaryHexer/nbw/internal/io/stream"
)

func NewDiodeWriter(w io.Writer, size int, poolInterval time.Duration, f diode.Alerter) *diode.Writer {
	return diode.NewWriter(w, size, poolInterval, f)
}

func NewBundlerWriter(w io.Writer, opts ...bundler.WriterOption) *bundler.Writer {
	return bundler.NewWriter(w, opts)
}

func NewStreamWriter(w io.Writer, flows ...streams.Flow) *stream.Writer {
	return stream.NewWriter(w, flows)
}
