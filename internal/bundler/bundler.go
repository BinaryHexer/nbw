package bundler

import (
	bbx "github.com/BinaryHexer/nbw/pkg/bundler"
	"google.golang.org/api/support/bundler"
	"time"
)

const (
	DefaultDelayThreshold       = time.Second
	DefaultBundleCountThreshold = 1000
	DefaultBundleByteThreshold  = 1e6     // 1MiB
	DefaultBundleByteLimit      = 0       // unlimited
	DefaultBufferedByteLimit    = 8 * 1e6 // 8MiB
)

func NewBundler(itemExample interface{}, handler func(interface{}), opts ...bbx.Option) *bundler.Bundler {
	b := bundler.NewBundler(itemExample, handler)

	b.DelayThreshold = DefaultDelayThreshold
	b.BundleCountThreshold = DefaultBundleCountThreshold
	b.BundleByteThreshold = DefaultBundleByteThreshold
	b.BundleByteLimit = DefaultBundleByteLimit
	b.BufferedByteLimit = DefaultBufferedByteLimit

	for _, o := range opts {
		o(b)
	}

	return b
}
