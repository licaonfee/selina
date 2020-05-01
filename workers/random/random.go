package random

import (
	"context"
	"crypto/rand"

	"github.com/licaonfee/selina"
)

var _ selina.Worker = (*Random)(nil)

//Options to customize Ramdom generator
type Options struct {
	//Len how many random bytes will contain []byte
	Len int
}

//Random generate random []byte slices
type Random struct {
	opts Options
}

//Process implements Worker interface
func (r *Random) Process(ctx context.Context, args selina.ProcessArgs) error {
	defer close(args.Output)
	for {
		msg := make([]byte, r.opts.Len)
		if _, err := rand.Read(msg); err != nil {
			return err
		}
		select {
		case args.Output <- msg:
		case _, ok := <-args.Input:
			if !ok {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

//NewRandom create a new Random worker that generate random byte slices
func NewRandom(opts Options) *Random {
	return &Random{opts: opts}
}
