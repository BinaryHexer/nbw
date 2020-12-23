package benchmark

import (
	"fmt"
	"io"
	"io/ioutil"
	"testing"
	"time"

	"github.com/BinaryHexer/nbw"
	"github.com/BinaryHexer/nbw/internal/io/bundler"
	"github.com/reugn/go-streams/flow"
)

func BenchmarkDiscardWriter(b *testing.B) {
	b.Logf("Writing a random string")

	b.Run("ioutil.Discard", func(b *testing.B) {
		w := discardWriter()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = w.Write([]byte(getMessage()))
			}
		})
	})
	b.Run("Diode", func(b *testing.B) {
		fw := fileWriter()
		w := nbw.NewDiodeWriter(fw, b.N, time.Second, func(missed int) {
			fmt.Printf("Logger Dropped %d messages", missed)
		})
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = w.Write([]byte(getMessage()))
			}
		})
	})
	b.Run("Bundler", func(b *testing.B) {
		fw := fileWriter()
		w := nbw.NewBundlerWriter(fw, bundler.WithBufferedByteLimit(b.N*1024))
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = w.Write([]byte(getMessage()))
			}
		})
	})
	b.Run("Diode.Bundler", func(b *testing.B) {
		fw := fileWriter()
		bw := nbw.NewBundlerWriter(fw, bundler.WithBufferedByteLimit(b.N*1024))
		w := nbw.NewDiodeWriter(bw, b.N, time.Second, func(missed int) {
			fmt.Printf("Logger Dropped %d messages", missed)
		})
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = w.Write([]byte(getMessage()))
			}
		})
	})
	b.Run("Bundler.Diode", func(b *testing.B) {
		fw := fileWriter()
		dw := nbw.NewDiodeWriter(fw, b.N, time.Second, func(missed int) {
			fmt.Printf("Logger Dropped %d messages", missed)
		})
		w := nbw.NewBundlerWriter(dw, bundler.WithBufferedByteLimit(b.N*1024))
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = w.Write([]byte(getMessage()))
			}
		})
	})
	b.Run("Stream", func(b *testing.B) {
		fw := fileWriter()
		w := nbw.NewStreamWriter(fw, flow.NewPassThrough())
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = w.Write([]byte(getMessage()))
			}
		})
	})
}

func BenchmarkFileWriter(b *testing.B) {
	b.Logf("Writing a random string")

	b.Run("os.File", func(b *testing.B) {
		w := fileWriter()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = w.Write([]byte(getMessage()))
			}
		})
	})
	b.Run("Diode", func(b *testing.B) {
		fw := fileWriter()
		w := nbw.NewDiodeWriter(fw, b.N, time.Second, func(missed int) {
			fmt.Printf("Logger Dropped %d messages", missed)
		})
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = w.Write([]byte(getMessage()))
			}
		})
	})
	b.Run("Bundler", func(b *testing.B) {
		fw := fileWriter()
		w := nbw.NewBundlerWriter(fw, bundler.WithBufferedByteLimit(b.N*1024))
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = w.Write([]byte(getMessage()))
			}
		})
	})
	b.Run("Diode.Bundler", func(b *testing.B) {
		fw := fileWriter()
		bw := nbw.NewBundlerWriter(fw, bundler.WithBufferedByteLimit(b.N*1024))
		w := nbw.NewDiodeWriter(bw, b.N, time.Second, func(missed int) {
			fmt.Printf("Logger Dropped %d messages", missed)
		})
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = w.Write([]byte(getMessage()))
			}
		})
	})
	b.Run("Bundler.Diode", func(b *testing.B) {
		fw := fileWriter()
		dw := nbw.NewDiodeWriter(fw, b.N, time.Second, func(missed int) {
			fmt.Printf("Logger Dropped %d messages", missed)
		})
		w := nbw.NewBundlerWriter(dw, bundler.WithBufferedByteLimit(b.N*1024))
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = w.Write([]byte(getMessage()))
			}
		})
	})
	b.Run("Stream", func(b *testing.B) {
		fw := fileWriter()
		w := nbw.NewStreamWriter(fw, flow.NewPassThrough())
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = w.Write([]byte(getMessage()))
			}
		})
	})
}

func discardWriter() io.Writer {
	return ioutil.Discard
}

func fileWriter() io.Writer {
	w, _ := ioutil.TempFile("/tmp", "writer_test*")
	return w
}
