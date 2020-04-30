package custom

import (
	"context"
	"errors"

	"github.com/licaonfee/selina"
)

var _ selina.Worker = (*Function)(nil)

//UserFunction define an user custom modification
//is safe to return input to avoid allocations
//if an error is returned Process is aborted
//a filter can be implemented returning (nil,nil)
type UserFunction func(input []byte) ([]byte, error)

//FunctionOptions customize a Function Worker
type FunctionOptions struct {
	Func UserFunction
}

//Function allow users to create custom Workers just with a function
type Function struct {
	opts FunctionOptions
}

//ErrNilFunction a nil UserFunction is provided via FunctionOptions
var ErrNilFunction = errors.New("nil UserFunction passed to Worker")

//Process implements selina.Workers
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
			if omsg == nil {
				continue
			}
			if err := selina.SendContext(ctx, omsg, output); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

//NewFunction create a Function object with goven options
func NewFunction(opts FunctionOptions) *Function {
	return &Function{opts: opts}
}