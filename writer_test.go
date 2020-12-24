package nbw

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/reugn/go-streams"
	"github.com/stretchr/testify/assert"

	"fmt"
	"github.com/BinaryHexer/nbw/pkg/stream"
	"io"
	"sync"
	"time"
)

func TestBundlerWriter(t *testing.T) {
	tests := []struct {
		msg string
	}{
		{msg: "Hello, World!"},
		{msg: "1234567890"},
		{msg: "@#$^%&*()!~"},
		{msg: `{"uuid":"ID001","level":"info","request":{"int":6,"float":7.19,"str":"apple","str_arr":["a","b"]}`},
	}

	for _, tt := range tests {
		buf := &bytes.Buffer{}
		w := NewBundlerWriter(buf)

		_, err := w.Write([]byte(tt.msg))

		assert.NoError(t, err)
		err = w.Close()
		assert.NoError(t, err)
		assert.Equal(t, tt.msg, buf.String())
	}
}

func TestDiodeWriter(t *testing.T) {
	tests := []struct {
		msg string
	}{
		{msg: "Hello, World!"},
		{msg: "1234567890"},
		{msg: "@#$^%&*()!~"},
		{msg: `{"uuid":"ID001","level":"info","request":{"int":6,"float":7.19,"str":"apple","str_arr":["a","b"]}`},
	}

	for _, tt := range tests {
		buf := &bytes.Buffer{}
		w := NewDiodeWriter(buf, 1000, 0, func(missed int) {})

		_, err := w.Write([]byte(tt.msg))

		assert.NoError(t, err)
		err = w.Close()
		assert.NoError(t, err)
		assert.Equal(t, tt.msg, buf.String())
	}
}

func TestStreamWriter(t *testing.T) {
	tests := []struct {
		msgs  []string
		want  []string
		flows []streams.Flow
	}{
		{
			msgs:  []string{"Hello, World!"},
			want:  []string{"Hello, World!"},
			flows: []streams.Flow{},
		},
		{
			msgs:  []string{"1234567890"},
			want:  []string{"1234567890"},
			flows: []streams.Flow{},
		},
		{
			msgs:  []string{"@#$^%&*()!~"},
			want:  []string{"@#$^%&*()!~"},
			flows: []streams.Flow{},
		},
		{
			msgs: []string{
				`{"uuid":"ID001","level":"debug","path":"/api/path","request_size":1100}`,
				`{"uuid":"ID001","level":"info","path":"/api/path","request":{"int":6,"float":7.19,"str":"apple","str_arr":["a","b"]},"msg":"request"}`,
				`{"uuid":"ID001","level":"error","msg":"error occurred"}`,
			},
			want: []string{
				`{"uuid":"ID001","level":"info","path":"/api/path","request":{"int":6,"float":7.19,"str":"apple","str_arr":["a","b"]},"msg":"request"}`,
				`{"uuid":"ID001","level":"error","msg":"error occurred"}`,
			},
			flows: []streams.Flow{
				stream.NewBasicFlow(
					func(i []byte) (stream.Metadata, []byte) {
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

						path, ok := obj["path"]
						if !ok {
							path = ""
						}

						md = stream.Metadata{
							"uuid":  uuid.(string),
							"level": level.(string),
							"path":  path.(string),
						}

						return md, i
					},
					func(md stream.Metadata) bool {
						l := md["level"]
						if l == "debug" {
							return false
						}

						return true
					},
					func(md stream.Metadata) string {
						return md["uuid"]
					},
					func(mds []stream.Metadata) bool {
						return true
					},
				),
			},
		},
	}

	for _, tt := range tests {
		buf := &bytes.Buffer{}
		w := NewStreamWriter(buf, tt.flows...)

		for _, msg := range tt.msgs {
			_, err := w.Write([]byte(msg + "\n"))
			assert.NoError(t, err)
		}

		err := w.Close()
		assert.NoError(t, err)

		got := strings.Split(buf.String(), "\n")
		want := append(tt.want, "")
		assert.ElementsMatch(t, want, got)
	}
}

func TestWriterConcurrent(t *testing.T) {
	const msgFormat = "Hello World, %d\n"
	const writes = 1000
	const delta = 0.01

	tests := []struct {
		w func(w io.Writer) io.WriteCloser
	}{
		{w: func(w io.Writer) io.WriteCloser { return NewBundlerWriter(w) }},
		{w: func(w io.Writer) io.WriteCloser { return NewDiodeWriter(w, 1000, 0, func(missed int) {}) }},
		{w: func(w io.Writer) io.WriteCloser { return NewStreamWriter(w) }},
	}

	for idx, tt := range tests {
		idx := idx
		tt := tt

		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			buf := &syncBuffer{
				buffer: bytes.Buffer{},
				mutex:  sync.Mutex{},
			}
			w := tt.w(buf)

			go func() {
				for i := 0; i < writes/2; i++ {
					msg := fmt.Sprintf(msgFormat, i)
					_, err := w.Write([]byte(msg))
					assert.NoError(t, err)
				}
			}()

			go func() {
				for i := writes / 2; i < writes; i++ {
					msg := fmt.Sprintf(msgFormat, i)
					_, err := w.Write([]byte(msg))
					assert.NoError(t, err)
				}
			}()

			done := make(chan bool)
			go func() {
				time.Sleep(500 * time.Millisecond)
				err := w.Close()
				assert.NoError(t, err)
				done <- true
			}()
			<-done

			_, _ = w.Write([]byte("write after closing"))

			lines := strings.Split(buf.String(), "\n")
			actual := float64(len(lines)-1) / writes
			assert.LessOrEqual(t, actual, 1.00)
			assert.InDelta(t, 1, actual, delta)
		})
	}
}

// Buffer is a goroutine safe bytes.Buffer
type syncBuffer struct {
	buffer bytes.Buffer
	mutex  sync.Mutex
}

// Write appends the contents of p to the buffer, growing the buffer as needed. It returns
// the number of bytes written.
func (s *syncBuffer) Write(p []byte) (n int, err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.buffer.Write(p)
}

// String returns the contents of the unread portion of the buffer
// as a string.  If the Buffer is a nil pointer, it returns "<nil>".
func (s *syncBuffer) String() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.buffer.String()
}
