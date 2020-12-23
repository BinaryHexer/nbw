package bundler

import (
	"google.golang.org/api/support/bundler"
	"time"
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
