package main

import (
	"encoding/json"
	"github.com/BinaryHexer/nbw"
	iox "github.com/BinaryHexer/nbw/pkg/io"
	"github.com/BinaryHexer/nbw/pkg/stream"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
	"sync"
	"time"
)

func main() {
	flow := stream.NewBasicFlow(basicFlow())
	writer := nbw.NewStreamWriter(os.Stdout, flow)
	logger := newNonBlockingZapLogger(zapcore.DebugLevel, writer)

	var wg sync.WaitGroup

	// all of these logs should be filtered out from the final output due to filterFunc.
	go func() {
		wg.Add(1)
		for i := 0; i < 10; i++ {
			iter := i
			logger.Error("hello", zap.String("uuid", "ID001"), zap.Int("iter", iter))
		}
		wg.Done()
	}()

	// all of these logs should be filtered out from the final output due to the groupFilterFunc.
	go func() {
		wg.Add(1)
		for i := 0; i < 10; i++ {
			iter := i
			logger.Info("hello", zap.String("uuid", "ID002"), zap.Int("iter", iter))
		}
		wg.Done()
	}()

	// all of these logs should be present in the final output due to last error log.
	go func() {
		wg.Add(1)
		for i := 0; i < 10; i++ {
			iter := i
			logger.Info("hello", zap.String("uuid", "ID003"), zap.Int("iter", iter))
		}
		logger.Error("hello", zap.String("uuid", "ID003"), zap.Int("iter", 10))
		wg.Done()
	}()

	wg.Wait()

	time.Sleep(100 * time.Millisecond)
	_ = logger.Sync()
}

func basicFlow() (stream.MapFn, stream.FilterFn, stream.GroupFn, stream.GroupFilterFn) {
	// extract metadata from logs
	mapFn := func(i []byte) (stream.Metadata, []byte) {
		var obj map[string]interface{}
		md := make(stream.Metadata)

		err := json.Unmarshal(i, &obj)
		if err != nil {
			return md, i
		}

		uuid, ok := obj["uuid"]
		if !ok {
			uuid = ""
		}

		level, ok := obj["level"]
		if !ok {
			level = ""
		}

		md = stream.Metadata{
			"uuid":  uuid.(string),
			"level": level.(string),
		}

		return md, i
	}

	// remove any logs with uuid ID001
	filterFn := func(md stream.Metadata) bool {
		uuid := md["uuid"]
		if uuid == "ID001" {
			return false
		}

		return true
	}

	// group by uuid
	groupFn := func(md stream.Metadata) string {
		return md["uuid"]
	}

	// remove any group with <1 error log
	groupFilterFn := func(mds []stream.Metadata) bool {
		for _, md := range mds {
			l := md["level"]
			if l == "error" {
				return true
			}
		}

		return false
	}

	return mapFn, filterFn, groupFn, groupFilterFn
}

func newNonBlockingZapLogger(lvl zapcore.Level, w io.Writer) *zap.Logger {
	ec := zap.NewProductionEncoderConfig()
	ec.EncodeDuration = zapcore.NanosDurationEncoder
	ec.EncodeTime = zapcore.EpochNanosTimeEncoder
	enc := zapcore.NewJSONEncoder(ec)
	return zap.New(zapcore.NewCore(
		enc,
		iox.AddCloserSync(w),
		lvl,
	))
}
