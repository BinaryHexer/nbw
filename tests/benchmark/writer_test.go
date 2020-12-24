package benchmark

import (
	"fmt"
	"io"
	"io/ioutil"
	"testing"
	"time"

	"encoding/json"
	"github.com/BinaryHexer/nbw"
	"github.com/BinaryHexer/nbw/internal/io/bundler"
	"github.com/BinaryHexer/nbw/pkg/stream"
	"github.com/reugn/go-streams/flow"
	"math/rand"
)

func BenchmarkWriter(b *testing.B) {
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

func BenchmarkStreamWriter(b *testing.B) {
	b.Run("BasicFlow.Stream", func(b *testing.B) {
		fw := fileWriter()
		w := nbw.NewStreamWriter(fw, stream.NewBasicFlow(basicFlow()))
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				msg := fmt.Sprintf(`{"uuid":"%d","level":"info","path":"/api/path","request":{"int":6,"float":7.19,"str":"apple","str_arr":["a","b"]},"msg":"request"}`, rand.Int())
				_, _ = w.Write([]byte(msg))
			}
		})
	})

	b.Run("BasicFlow.Diode.Stream", func(b *testing.B) {
		fw := fileWriter()
		sw := nbw.NewStreamWriter(fw, stream.NewBasicFlow(basicFlow()))
		w := nbw.NewDiodeWriter(sw, b.N, time.Second, func(missed int) {
			fmt.Printf("Logger Dropped %d messages", missed)
		})
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				msg := fmt.Sprintf(`{"uuid":"%d","level":"info","path":"/api/path","request":{"int":6,"float":7.19,"str":"apple","str_arr":["a","b"]},"msg":"request"}`, rand.Int())
				_, _ = w.Write([]byte(msg))
			}
		})
	})

	b.Run("BasicFlow.Bundler.Stream", func(b *testing.B) {
		fw := fileWriter()
		sw := nbw.NewStreamWriter(fw, stream.NewBasicFlow(basicFlow()))
		w := nbw.NewBundlerWriter(sw, bundler.WithBufferedByteLimit(b.N*1024))
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				msg := fmt.Sprintf(`{"uuid":"%d","level":"info","path":"/api/path","request":{"int":6,"float":7.19,"str":"apple","str_arr":["a","b"]},"msg":"request"}`, rand.Int())
				_, _ = w.Write([]byte(msg))
			}
		})
	})
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

		md = stream.Metadata{
			"uuid": uuid.(string),
		}

		i, _ = json.Marshal(obj)
		i = append(i, '\n')

		return md, i
	}

	filterFn := func(md stream.Metadata) bool {
		return true
	}

	groupFn := func(md stream.Metadata) string {
		return md["uuid"]
	}

	groupFilterFn := func(mds []stream.Metadata) bool {
		return true
	}

	return mapFn, filterFn, groupFn, groupFilterFn
}

func discardWriter() io.Writer {
	return ioutil.Discard
}

func fileWriter() io.Writer {
	w, _ := ioutil.TempFile("/tmp", "writer_test*")
	return w
}
