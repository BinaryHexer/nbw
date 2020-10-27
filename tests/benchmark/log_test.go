// https://github.com/uber-go/zap/blob/master/benchmarks/scenario_bench_test.go
package benchmark

import (
	"flag"
	"fmt"
	iox "io"
	"io/ioutil"
	"testing"
	"time"

	"github.com/BinaryHexer/nbw"
	"go.uber.org/zap"
)

var fw = flag.String("writer", "discard", "os writer to use: discard or file")

func BenchmarkWithoutFields(b *testing.B) {
	b.Logf("Logging without any structured context.")
	b.Run("Zap", func(b *testing.B) {
		logger := newZapLogger(zap.DebugLevel)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage())
			}
		})
	})
	b.Run("Zap.Bundler", func(b *testing.B) {
		w := nbw.NewBundlerWriter(newWriter())
		logger := newNonBlockingZapLogger(zap.DebugLevel, w)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage())
			}
		})
	})
	b.Run("Zap.Diode", func(b *testing.B) {
		w := nbw.NewDiodeWriter(newWriter(), b.N, 10*time.Millisecond, func(missed int) {
			fmt.Printf("Logger Dropped %d messages", missed)
		})
		logger := newNonBlockingZapLogger(zap.DebugLevel, w)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage())
			}
		})
	})
	b.Run("Zap.Diode.Bundler", func(b *testing.B) {
		bw := nbw.NewBundlerWriter(newWriter())
		w := nbw.NewDiodeWriter(bw, b.N, 10*time.Millisecond, func(missed int) {
			fmt.Printf("Logger Dropped %d messages", missed)
		})
		logger := newNonBlockingZapLogger(zap.DebugLevel, w)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage())
			}
		})
	})
}

func BenchmarkAccumulatedContext(b *testing.B) {
	b.Logf("Logging with some accumulated context.")
	b.Run("Zap", func(b *testing.B) {
		logger := newZapLogger(zap.DebugLevel).With(fakeFields()...)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage())
			}
		})
	})
	b.Run("Zap.Bundler", func(b *testing.B) {
		w := nbw.NewBundlerWriter(newWriter())
		logger := newNonBlockingZapLogger(zap.DebugLevel, w).With(fakeFields()...)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage())
			}
		})
	})
	b.Run("Zap.Diode", func(b *testing.B) {
		w := nbw.NewDiodeWriter(newWriter(), b.N, 10*time.Millisecond, func(missed int) {
			fmt.Printf("Logger Dropped %d messages", missed)
		})
		logger := newNonBlockingZapLogger(zap.DebugLevel, w).With(fakeFields()...)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage())
			}
		})
	})
	b.Run("Zap.Diode.Bundler", func(b *testing.B) {
		bw := nbw.NewBundlerWriter(newWriter())
		w := nbw.NewDiodeWriter(bw, b.N, 10*time.Millisecond, func(missed int) {
			fmt.Printf("Logger Dropped %d messages", missed)
		})
		logger := newNonBlockingZapLogger(zap.DebugLevel, w).With(fakeFields()...)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage())
			}
		})
	})
}

func BenchmarkAddingFields(b *testing.B) {
	b.Logf("Logging with additional context at each log site.")
	b.Run("Zap", func(b *testing.B) {
		logger := newZapLogger(zap.DebugLevel)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(), fakeFields()...)
			}
		})
	})
	b.Run("Zap.Bundler", func(b *testing.B) {
		w := nbw.NewBundlerWriter(newWriter())
		logger := newNonBlockingZapLogger(zap.DebugLevel, w)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(), fakeFields()...)
			}
		})
	})
	b.Run("Zap.Diode", func(b *testing.B) {
		wr := nbw.NewDiodeWriter(newWriter(), b.N, 10*time.Millisecond, func(missed int) {
			fmt.Printf("Logger Dropped %d messages", missed)
		})
		logger := newNonBlockingZapLogger(zap.DebugLevel, wr)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(), fakeFields()...)
			}
		})
	})
	b.Run("Zap.Diode.Bundler", func(b *testing.B) {
		bw := nbw.NewBundlerWriter(newWriter())
		w := nbw.NewDiodeWriter(bw, b.N, 10*time.Millisecond, func(missed int) {
			fmt.Printf("Logger Dropped %d messages", missed)
		})
		logger := newNonBlockingZapLogger(zap.DebugLevel, w)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(), fakeFields()...)
			}
		})
	})
}

func newWriter() iox.Writer {
	flag.Parse()
	f := *fw
	if f == "discard" {
		return ioutil.Discard
	} else {
		w, _ := ioutil.TempFile("/tmp", "logtest*")
		return w
	}
}
