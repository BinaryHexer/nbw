package main

import (
	"github.com/BinaryHexer/nbw"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
)

func main() {
	// can be replaced with NewDiodeWrite or NewStreamWriter
	writer := nbw.NewBundlerWriter(os.Stdout)
	logger := newNonBlockingZapLogger(zapcore.InfoLevel, writer)

	logger.Info("hello world", zap.String("a", "1"))

	_ = logger.Sync()
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
