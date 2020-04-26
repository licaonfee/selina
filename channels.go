package selina

import (
	"sync"
)

//Broadcaster allow to write same value to multiple groutines
type Broadcaster struct {
	out []chan<- []byte
	mx  sync.RWMutex
}

//Broadcast read values from input and send it to output channels
func (b *Broadcaster) Broadcast(input <-chan []byte) {
	for in := range input {
		b.mx.RLock()
		for _, out := range b.out {
			data := make([]byte, len(in))
			copy(data, in)
			//TODO: this can be a deadlock
			out <- data
		}
		b.mx.RUnlock()
	}
	//close all channels when all data is readed
	for _, c := range b.out {
		close(c)
	}
}

//Client create an output chanel, it returns an error if Broadcast is already called
func (b *Broadcaster) Client() <-chan []byte {
	b.mx.RLock()
	defer b.mx.RUnlock()
	c := make(chan []byte)
	b.out = append(b.out, c)
	return c
}

//Receiver join multiple channels into a single output channel
// this allow to add new channels after Receive is called
type Receiver struct {
	out chan []byte
	wg  sync.WaitGroup
	// this is used to always initialize channel and allow to use receiver with default value
	init sync.Once
}

func (r *Receiver) pipe(in <-chan []byte) {
	for msg := range in {
		r.out <- msg
	}
	r.wg.Done()
}

func (r *Receiver) initChan() {
	r.init.Do(func() {
		r.out = make(chan []byte)
	})
}

//Receive listen to all channels configured with Watch
// when all channels are closed, output chanel is closed too
func (r *Receiver) Receive() <-chan []byte {
	r.initChan()
	go func() {
		r.wg.Wait()
		close(r.out)
	}()
	return r.out
}

//Watch add a new channel to be joined
// Call Watch after Receive is a panic
func (r *Receiver) Watch(input <-chan []byte) {
	r.initChan()
	r.wg.Add(1)
	go r.pipe(input)
}
