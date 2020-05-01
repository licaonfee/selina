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

func (l *lazyWorker) Process(ctx context.Context, args selina.ProcessArgs) error {
	for {
		select {
		case _, ok := <-args.Input:
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

func (p *produceN) Process(ctx context.Context, args selina.ProcessArgs) error {
	for i := 0; i < p.count; i++ {
		args.Output <- p.message
	}
	close(args.Output)
	return nil
}

type sink struct{}

func (s *sink) Process(ctx context.Context, args selina.ProcessArgs) error {
	for range args.Input {
	}
	close(args.Output)
	return nil
}

type dummyWorker struct{}

func (d *dummyWorker) Process(ctx context.Context, args selina.ProcessArgs) error {
	for msg := range args.Input {
		args.Output <- msg
	}
	return nil
}

type sliceReader struct {
	values []string
}

func (r *sliceReader) Process(ctx context.Context, args selina.ProcessArgs) error {
	rd := selina.SliceAsChannel(r.values, true)
	for msg := range rd {
		args.Output <- msg
	}
	return nil
}

type sliceWriter struct {
	values []string
}

func (w *sliceWriter) Process(ctx context.Context, args selina.ProcessArgs) error {
	for msg := range args.Input {
		w.values = append(w.values, string(msg))
	}

	return nil
}
