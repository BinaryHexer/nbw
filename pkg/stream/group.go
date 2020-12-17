package stream

import (
	bbx "github.com/BinaryHexer/nbw/internal/bundler"
	"github.com/reugn/go-streams"
	"google.golang.org/api/support/bundler"
	"reflect"
	"sync"
	"time"
)

const (
	DefaultDelayThreshold       = 2 * time.Second
	DefaultBundleCountThreshold = 1000
	DefaultBufferedByteLimit    = 10 * 1e6 // 10MiB
)

// GroupFunc is a filter predicate function.
type GroupFunc func(interface{}) string

type wrappedElement struct {
	key  string
	data interface{}
}

// Aggregator groups the incoming elements using a function.
// The elements are grouped by the key returned by the function.
//
//   eg:
//    GroupF(1,2,3) = a
//    GroupF(4,5)   = b
//
// in  -- 1 -- 2 ---- 3 -- 4 ------ 5 --
//        |    |      |    |        |
//    [---------- AggregatorFunc --------]
//                    |               |
// out --------------[1,2,3]---------[4,5] --
type Aggregator struct {
	GroupF      GroupFunc
	in          chan interface{}
	out         chan interface{}
	evict       chan string
	done        chan struct{}
	bundlers    map[string]*bundler.Bundler
	bundlerPool *sync.Pool
}

// NewAggregator returns a new Aggregator instance.
// groupFunc is the grouping function.
func NewAggregator(groupFunc GroupFunc, opts ...bbx.Option) *Aggregator {
	a := &Aggregator{
		GroupF:   groupFunc,
		in:       make(chan interface{}),
		out:      make(chan interface{}),
		evict:    make(chan string, 10),
		done:     make(chan struct{}),
		bundlers: make(map[string]*bundler.Bundler),
	}
	pool := &sync.Pool{
		New: func() interface{} {
			return a.newBundler(opts...)
		},
	}
	a.bundlerPool = pool

	go a.gc()
	go a.receive()

	return a
}

// Via streams data through the given flow
func (a *Aggregator) Via(flow streams.Flow) streams.Flow {
	go a.transmit(flow)
	return flow
}

// To streams data to the given sink
func (a *Aggregator) To(sink streams.Sink) {
	a.transmit(sink)
}

// Out returns an output channel for sending data
func (a *Aggregator) Out() <-chan interface{} {
	return a.out
}

// In returns an input channel for receiving data
func (a *Aggregator) In() chan<- interface{} {
	return a.in
}

func (a *Aggregator) transmit(inlet streams.Inlet) {
	for elem := range a.Out() {
		inlet.In() <- elem
	}
	close(inlet.In())
}

// groups elements using the group function
func (a *Aggregator) receive() {
	// read from the input channel
	for elem := range a.in {
		a.store(a.GroupF(elem), elem)
	}

	// flush all bundlers
	for k := range a.bundlers {
		a.evict <- k
	}

	// close the evict channel
	close(a.evict)

	// wait for evict channel to be consumed
	<-a.done

	// close the output channel
	close(a.out)
}

func (a *Aggregator) store(k string, e interface{}) {
	we := &wrappedElement{key: k, data: e}
	b := a.getBundler(k)

	size := reflect.TypeOf(we).Size()
	err := b.Add(we, int(size))
	if err != nil {
		panic(err)
	}
}

func (a *Aggregator) getBundler(k string) *bundler.Bundler {
	if k == "" {
		k = "default"
	}

	b, ok := a.bundlers[k]
	if !ok {
		b = a.bundlerPool.Get().(*bundler.Bundler)
		a.bundlers[k] = b
	}

	return b
}

func (a *Aggregator) newBundler(opts ...bbx.Option) *bundler.Bundler {
	var e wrappedElement
	b := bbx.NewBundler(&e, func(p interface{}) {
		a.emit(p.([]*wrappedElement))
	})

	b.BundleCountThreshold = DefaultBundleCountThreshold
	b.DelayThreshold = DefaultDelayThreshold
	b.BufferedByteLimit = DefaultBufferedByteLimit

	for _, o := range opts {
		o(b)
	}

	return b
}

func (a *Aggregator) emit(elements []*wrappedElement) {
	t := make([]interface{}, len(elements))
	k := ""

	for idx, e := range elements {
		ex := *e
		k = ex.key
		t[idx] = ex.data
	}

	a.out <- t

	{
		defer func() {
			if r := recover(); r != nil {
				// do nothing
			}
		}()
		a.evict <- k
	}
}

func (a *Aggregator) gc() {
	for k := range a.evict {
		b, ok := a.bundlers[k]
		if ok {
			// remove from map so no additional data is written
			delete(a.bundlers, k)
			// ensure all data in the bundler is flushed
			b.Flush()
			// return bundler to the pool
			a.bundlerPool.Put(b)
		}
	}
	close(a.done)
}
