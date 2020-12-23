package stream

import (
	ext "github.com/reugn/go-streams/extension"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAggregator(t *testing.T) {
	in := make(chan interface{})
	out := make(chan interface{})
	done := make(chan bool)

	aggr := NewAggregator(func(i interface{}) string {
		e := i.(int)
		if e%2 == 0 {
			return "even"
		}

		return "odd"
	})

	source := ext.NewChanSource(in)
	flow := aggr
	sink := ext.NewChanSink(out)

	var _input = []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	var _expectedOutput = [][]int{{1, 3, 5, 7, 9}, {2, 4, 6, 8}}

	go ingest(_input, in, done)
	go deferClose(done, in)
	go func() {
		source.
			Via(flow).
			To(sink)
	}()

	var _output [][]int
	for e := range sink.Out {
		xs := e.([]interface{})
		tmp := make([]int, len(xs))
		for idx, x := range xs {
			tmp[idx] = x.(int)
		}
		_output = append(_output, tmp)
	}
	assert.ElementsMatch(t, _expectedOutput, _output)
}

func ingest(source []int, in chan interface{}, done chan bool) {
	for _, e := range source {
		in <- e
	}
	done <- true
}

func deferClose(done chan bool, in chan interface{}) {
	<-done
	close(in)
}
