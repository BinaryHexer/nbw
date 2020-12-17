package main

import (
	"github.com/BinaryHexer/nbw"
	"github.com/reugn/go-streams/flow"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/time/rate"
	"io"
	"os"
	"time"
)

func main() {
	limiter := rate.NewLimiter(rate.Every(time.Second/10), 1)
	samplingFn := func(i interface{}) bool {
		return limiter.Allow()
	}

	writer := nbw.NewStreamWriter(os.Stdout, flow.NewFilter(samplingFn, 1))
	logger := newNonBlockingZapLogger(zapcore.InfoLevel, writer)

	// 20 logs / sec - 50% sampling
	for i := 0; i < 15; i++ {
		iter := i
		logger.Info("hello world", zap.Int("iter", iter))
		time.Sleep(50 * time.Millisecond)
	}

	time.Sleep(time.Second)

	// 10 logs / sec - 0% sampling
	for i := 15; i < 25; i++ {
		iter := i
		logger.Info("hello world", zap.Int("iter", iter))
		time.Sleep(100 * time.Millisecond)
	}
}

func newNonBlockingZapLogger(lvl zapcore.Level, w io.Writer) *zap.Logger {
	ec := zap.NewProductionEncoderConfig()
	ec.EncodeDuration = zapcore.NanosDurationEncoder
	ec.EncodeTime = zapcore.EpochNanosTimeEncoder
	enc := zapcore.NewJSONEncoder(ec)
	return zap.New(zapcore.NewCore(
		enc,
		zapcore.AddSync(w),
		lvl,
	))
}
