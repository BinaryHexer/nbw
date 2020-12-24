package atomic

import "sync/atomic"

type Bool struct{ flag int32 }

func (b *Bool) Set(value bool) {
	var i int32 = 0
	if value {
		i = 1
	}
	atomic.StoreInt32(&(b.flag), i)
}

func (b *Bool) Get() bool {
	return atomic.LoadInt32(&(b.flag)) != 0
}
