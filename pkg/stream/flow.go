package stream

import (
	bbx "github.com/BinaryHexer/nbw/pkg/bundler"
	"github.com/reugn/go-streams"
	"github.com/reugn/go-streams/flow"
	"runtime"
)

type Metadata map[string]string

type msg struct {
	md Metadata
	d  interface{}
}

type xmsg struct {
	mds []Metadata
	d   interface{}
}

type MapFn func([]byte) (Metadata, []byte)
type FilterFn func(Metadata) bool
type GroupFn func(Metadata) string
type GroupFilterFn func([]Metadata) bool

func NewBasicFlow(mapFunc MapFn, filterFunc1 FilterFn, groupFunc GroupFn, filterFunc2 GroupFilterFn, opts ...bbx.Option) streams.Flow {
	p := uint(runtime.NumCPU())

	m1 := flow.NewMap(toMapFunc(mapFunc), p)
	f1 := flow.NewFilter(toFilterFunc(filterFunc1), p)
	a := NewAggregator(toGroupFunc(groupFunc), opts...)
	m2 := flow.NewMap(func(i interface{}) interface{} {
		xs := i.([]interface{})
		mds := make([]Metadata, len(xs))
		for idx, x := range xs {
			y := x.(*msg)
			mds[idx] = y.md
		}

		return &xmsg{
			mds: mds,
			d:   xs,
		}
	}, p)
	f2 := flow.NewFilter(toGroupFilterFunc(filterFunc2), p)
	fm := flow.NewFlatMap(func(i interface{}) []interface{} {
		y := i.(*xmsg)
		xs := y.d.([]interface{})
		is := make([]interface{}, len(xs))

		for idx, x := range xs {
			z := x.(*msg)
			is[idx] = z.d
		}

		return is
	}, p)

	return NewPipe(m1, f1, a, m2, f2, fm)
}

func toMapFunc(fn MapFn) flow.MapFunc {
	return func(i interface{}) interface{} {
		d := i.([]byte)
		md, d := fn(d)

		m := &msg{md: md, d: d}

		return m
	}
}

func toGroupFunc(fn GroupFn) GroupFunc {
	return func(i interface{}) string {
		x := i.(*msg)
		k := fn(x.md)

		return k
	}
}

func toFilterFunc(fn FilterFn) flow.FilterFunc {
	return func(i interface{}) bool {
		x := i.(*msg)
		r := fn(x.md)

		return r
	}
}

func toGroupFilterFunc(fn GroupFilterFn) flow.FilterFunc {
	return func(i interface{}) bool {
		x := i.(*xmsg)
		r := fn(x.mds)

		return r
	}
}
