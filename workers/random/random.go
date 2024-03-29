package random

import (
	"bytes"
	"context"
	"crypto/rand"
	"io"

	"github.com/licaonfee/selina"
)

var _ selina.Worker = (*Random)(nil)

// Options to customize Ramdom generator
type Options struct {
	// Len how many random bytes will contain []byte
	Len int
}

// Random generate random []byte slices
type Random struct {
	opts Options
}

// Process implements Worker interface
func (r *Random) Process(ctx context.Context, args selina.ProcessArgs) error {
	defer close(args.Output)

	isNil := args.Input == nil
	var input <-chan *bytes.Buffer
	if !isNil {
		input = args.Input
	} else {
		inp := make(chan *bytes.Buffer)
		close(inp)
		input = inp
	}
	for {
		select {
		case _, ok := <-input:
			if !ok && !isNil {
				return nil
			}
			msg := selina.GetBuffer()

			if _, err := io.CopyN(msg, rand.Reader, int64(r.opts.Len)); err != nil {
				return err
			}
			if err := selina.SendContext(ctx, msg, args.Output); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// NewRandom create a new Random worker that generate random byte slices
func NewRandom(opts Options) *Random {
	return &Random{opts: opts}
}
