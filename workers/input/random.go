package input

import (
	"context"
	"crypto/rand"

	"github.com/licaonfee/selina"
)

var _ selina.Worker = (*Random)(nil)

type RandomOptions struct {
	Len int
}

type Random struct {
	opts RandomOptions
}

func (r *Random) Process(ctx context.Context, input <-chan []byte, output chan<- []byte) error {
	defer close(output)
	for {
		msg := make([]byte, r.opts.Len)
		if _, err := rand.Read(msg); err != nil {
			return err
		}
		select {
		case output <- msg:
		case _, ok := <-input:
			if !ok {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func NewRandom(opts RandomOptions) *Random {
	return &Random{opts: opts}
}
