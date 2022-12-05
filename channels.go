package selina

import (
	"sync"

	"context"
)

// Broadcaster allow to write same value to multiple groutines
type Broadcaster struct {
	DataCounter
	out     []chan<- []byte
	mtx     sync.Mutex
	running bool
}

// Broadcast read values from input and send it to output channels
func (b *Broadcaster) Broadcast(input <-chan []byte) {
	b.mtx.Lock()
	b.running = true
	b.mtx.Unlock()
	for in := range input {
		for _, out := range b.out {
			data := make([]byte, len(in))
			copy(data, in)
			//TODO: this can be a deadlock
			out <- data
			b.SumData(data)
		}
	}
	// close all channels when all data is readed
	for _, c := range b.out {
		close(c)
	}
}

// Client create an output chanel, it panics if Broadcast is already called
func (b *Broadcaster) Client() <-chan []byte {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	if b.running {
		panic("call Client after Broadcast")
	}
	c := make(chan []byte)
	b.out = append(b.out, c)
	return c
}

// Receiver join multiple channels into a single output channel
// this allow to add new channels after Receive is called
type Receiver struct {
	DataCounter
	out chan []byte
	wg  sync.WaitGroup
	// this is used to always initialize channel and allow to use receiver with default value
	init sync.Once
}

func (r *Receiver) pipe(in <-chan []byte) {
	for msg := range in {
		r.out <- msg
		r.SumData(msg)
	}
	r.wg.Done()
}

func (r *Receiver) initChan() {
	r.init.Do(func() {
		r.out = make(chan []byte)
	})
}

// Receive listen to all channels configured with Watch
// when all channels are closed, output chanel is closed too
// if there is no channels in watch list , this method returns
// a nil channel
func (r *Receiver) Receive() <-chan []byte {
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
func (r *Receiver) Watch(input <-chan []byte) {
	r.initChan()
	r.wg.Add(1)
	go r.pipe(input)
}

// SendContext try to send msg to output, it returns an error if
// context is canceled before msg is sent
func SendContext(ctx context.Context, msg []byte, output chan<- []byte) error {
	select {
	case output <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
