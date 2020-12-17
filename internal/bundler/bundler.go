package bundler

import (
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

// Option can be used to setup the bundler.
type Option func(*bundler.Bundler)

// WithDelayThreshold sets the interval at which the bundler is flushed.
// The default is DefaultDelayThreshold.
func WithDelayThreshold(delay time.Duration) Option {
	return Option(func(b *bundler.Bundler) {
		b.DelayThreshold = delay
	})
}

// WithBundleCountThreshold sets the max number of items after which the bundler is flushed.
// The default is DefaultBundleCountThreshold.
func WithBundleCountThreshold(n int) Option {
	return Option(func(b *bundler.Bundler) {
		b.BundleCountThreshold = n
	})
}

// WithBundleByteThreshold sets the max size of the bundle (in bytes) at which the
// bundler is flushed. The default is DefaultBundleByteThreshold.
func WithBundleByteThreshold(n int) Option {
	return Option(func(b *bundler.Bundler) {
		b.BundleByteThreshold = n
	})
}

// WithBundleByteLimit sets the maximum size of a bundle, in bytes.
// Zero means unlimited. The default is DefaultBundleByteLimit.
func WithBundleByteLimit(n int) Option {
	return Option(func(b *bundler.Bundler) {
		b.BundleByteLimit = n
	})
}

// WithBufferedByteLimit sets the maximum number of bytes that the Bundler will keep
// in memory before returning ErrOverflow. The default is DefaultBufferedByteLimit.
func WithBufferedByteLimit(n int) Option {
	return Option(func(b *bundler.Bundler) {
		b.BufferedByteLimit = n
	})
}

func NewBundler(itemExample interface{}, handler func(interface{}), opts ...Option) *bundler.Bundler {
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
