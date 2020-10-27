package stream

import (
	"github.com/reugn/go-streams"
	"github.com/reugn/go-streams/flow"
)

// Pipe collapses multiple flows into a single one.
// Pipe can be used to replace
//    Source.
//          Via(f1).
//          Via(f2).
//          Via(f3).
//          To(Sink)
// with
//		Source.
//	      	  Via(NewPipe(f1,f2,f3)).
//			  To(Sink)
//
type Pipe struct {
	in  chan<- interface{}
	out <-chan interface{}
}

// NewPipe returns a new Pipe instance.
func NewPipe(flows ...streams.Flow) *Pipe {
	flows = append([]streams.Flow{flow.NewPassThrough()}, flows...)
	fi := flows[0]
	fl := flows[len(flows)-1]

	fx := fi
	for i := 1; i < len(flows); i++ {
		fx = fx.Via(flows[i])
	}

	return &Pipe{
		fi.In(),
		fl.Out(),
	}
}

// Via streams data through the given flow
func (p *Pipe) Via(flow streams.Flow) streams.Flow {
	go p.transmit(flow)
	return flow
}

// To streams data to the given sink
func (p *Pipe) To(sink streams.Sink) {
	p.transmit(sink)
}

// Out returns an output channel for sending data
func (p *Pipe) Out() <-chan interface{} {
	return p.out
}

// In returns an input channel for receiving data
func (p *Pipe) In() chan<- interface{} {
	return p.in
}

func (p *Pipe) transmit(inlet streams.Inlet) {
	for elem := range p.Out() {
		inlet.In() <- elem
	}
	close(inlet.In())
}
