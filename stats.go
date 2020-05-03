package selina

import (
	"sync/atomic"
)

//DataCounter a simple atomic wrapper
type DataCounter struct {
	count int64
	data  int64
}

//SumData icrement count+1 and data + len(msg)
// while both values are incremented in an atomic way
// is posible to get inconsistent reads on call Stats
// while object is in use
func (c *DataCounter) SumData(msg []byte) {
	atomic.AddInt64(&c.count, 1)
	atomic.AddInt64(&c.data, int64(len(msg)))
}

func (c *DataCounter) Stats() (count int64, data int64) {
	return atomic.LoadInt64(&c.count), atomic.LoadInt64(&c.data)
}
