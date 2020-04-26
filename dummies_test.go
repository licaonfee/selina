package selina_test

import (
	"context"

	"github.com/licaonfee/selina"
)

var _ selina.Worker = (*lazyWorker)(nil)
var _ selina.Worker = (*dummyWorker)(nil)
var _ selina.Worker = (*produceN)(nil)
var _ selina.Worker = (*sink)(nil)
var _ selina.Worker = (*sliceReader)(nil)
var _ selina.Worker = (*sliceWriter)(nil)

//lazyWorker just wait until context is canceled, or in is closed
type lazyWorker struct{}

func (l *lazyWorker) Process(ctx context.Context, in <-chan []byte, out chan<- []byte) error {
	for {
		select {
		case _, ok := <-in:
			if !ok {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

type produceN struct {
	count   int
	message []byte
}

func (p *produceN) Process(ctx context.Context, input <-chan []byte, output chan<- []byte) error {
	for i := 0; i < p.count; i++ {
		output <- p.message
	}
	close(output)
	return nil
}

type sink struct{}

func (s *sink) Process(ctx context.Context, input <-chan []byte, output chan<- []byte) error {
	for range input {
	}
	close(output)
	return nil
}

type dummyWorker struct{}

func (d *dummyWorker) Process(ctx context.Context, in <-chan []byte, out chan<- []byte) error {
	for msg := range in {
		out <- msg
	}
	return nil
}

type sliceReader struct {
	values []string
}

func (r *sliceReader) Process(ctx context.Context, in <-chan []byte, out chan<- []byte) error {
	rd := selina.SliceAsChannel(r.values, true)
	for msg := range rd {
		out <- msg
	}
	return nil
}

type sliceWriter struct {
	values []string
}

func (w *sliceWriter) Process(ctx context.Context, in <-chan []byte, out chan<- []byte) error {
	for msg := range in {
		w.values = append(w.values, string(msg))
	}

	return nil
}
