package selina

import (
	"bytes"
	"sync"

	"context"
)

var pool = sync.Pool{New: func() any {
	return bytes.NewBuffer(nil)
}}

// GetBuffer returns a buffer from a pool of buffers
func GetBuffer() *bytes.Buffer {
	return pool.Get().(*bytes.Buffer)
}

// FreeBuffer calls Buffer.Reset and return buffer to the pool
func FreeBuffer(b *bytes.Buffer) {
	if b == nil {
		return
	}
	b.Reset()
	pool.Put(b)
}

// Broadcaster allow to write same value to multiple groutines
type Broadcaster struct {
	DataCounter
	out     []chan<- *bytes.Buffer
	mtx     sync.Mutex
	running bool
}

// Broadcast read values from input and send it to output channels
func (b *Broadcaster) Broadcast(input <-chan *bytes.Buffer) {
	b.mtx.Lock()
	b.running = true
	b.mtx.Unlock()
	for in := range input {
		for _, out := range b.out {
			data := GetBuffer()
			data.Write(in.Bytes())
			//TODO: this can be a deadlock
			b.SumData(data.Bytes())
			out <- data
		}
		FreeBuffer(in)
	}
	// close all channels when all data is readed
	for _, c := range b.out {
		close(c)
	}
}

// Client create an output chanel, it panics if Broadcast is already called
func (b *Broadcaster) Client() <-chan *bytes.Buffer {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	if b.running {
		panic("call Client after Broadcast")
	}
	c := make(chan *bytes.Buffer)
	b.out = append(b.out, c)
	return c
}

// Receiver join multiple channels into a single output channel
// this allow to add new channels after Receive is called
type Receiver struct {
	DataCounter
	out chan *bytes.Buffer
	wg  sync.WaitGroup
	// this is used to always initialize channel and allow to use receiver with default value
	init sync.Once
}

func (r *Receiver) pipe(in <-chan *bytes.Buffer) {
	for msg := range in {
		r.SumData(msg.Bytes())
		r.out <- msg
	}
	r.wg.Done()
}

func (r *Receiver) initChan() {
	r.init.Do(func() {
		r.out = make(chan *bytes.Buffer)
	})
}

// Receive listen to all channels configured with Watch
// when all channels are closed, output chanel is closed too
// if there is no channels in watch list , this method returns
// a nil channel
func (r *Receiver) Receive() <-chan *bytes.Buffer {
	go func() {
		if r.out != nil {
			r.wg.Wait()
			close(r.out)
		}
	}()
	return r.out
}

// Watch add a new channel to be joined
// Call Watch after Receive is a panic
func (r *Receiver) Watch(input <-chan *bytes.Buffer) {
	r.initChan()
	r.wg.Add(1)
	go r.pipe(input)
}

// SendContext try to send msg to output, it returns an error if
// context is canceled before msg is sent
func SendContext(ctx context.Context, msg *bytes.Buffer, output chan<- *bytes.Buffer) error {
	select {
	case output <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
