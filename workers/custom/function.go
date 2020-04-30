package custom

import (
	"context"
	"errors"

	"github.com/licaonfee/selina"
)

var _ selina.Worker = (*Function)(nil)

type UserFunction func(input []byte) ([]byte, error)

type FunctionOptions struct {
	Func UserFunction
}

type Function struct {
	opts FunctionOptions
}

var ErrNilFunction = errors.New("nil UserFunction passed to Worker")

func (f *Function) Process(ctx context.Context, input <-chan []byte, output chan<- []byte) error {
	defer close(output)
	if f.opts.Func == nil {
		return ErrNilFunction
	}
	for {
		select {
		case msg, ok := <-input:
			if !ok {
				return nil
			}
			omsg, err := f.opts.Func(msg)
			if err != nil {
				return err
			}
			if err := send(ctx, omsg, output); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func send(ctx context.Context, msg []byte, output chan<- []byte) error {
	select {
	case output <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

//NewFunction create a Function object with goven options
func NewFunction(opts FunctionOptions) *Function {
	return &Function{opts: opts}
}
