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
	defer close(args.Output)
	for i := 0; i < p.count; i++ {
		select {
		case args.Output <- p.message:
		case <-ctx.Done():
			return ctx.Err()
		case _, ok := <-args.Input:
			if !ok {
				return nil
			}
		}
	}
	return nil
}

type sink struct{}

func (s *sink) Process(ctx context.Context, args selina.ProcessArgs) error {
	defer close(args.Output)
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

type dummyWorker struct{}

func (d *dummyWorker) Process(ctx context.Context, args selina.ProcessArgs) error {
	defer close(args.Output)
	for {
		select {
		case msg, ok := <-args.Input:
			if !ok {
				return nil
			}
			if err := selina.SendContext(ctx, msg, args.Output); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

type sliceReader struct {
	values []string
}

func (r *sliceReader) Process(ctx context.Context, args selina.ProcessArgs) error {
	defer close(args.Output)
	for _, v := range r.values {
		select {
		default:
			if err := selina.SendContext(ctx, []byte(v), args.Output); err != nil {
				return err
			}
		case _, ok := <-args.Input:
			if !ok {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

type sliceWriter struct {
	values []string
}

func (w *sliceWriter) Process(ctx context.Context, args selina.ProcessArgs) error {
	defer close(args.Output)
	for {
		select {
		case msg, ok := <-args.Input:
			if !ok {
				return nil
			}
			w.values = append(w.values, string(msg))
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
