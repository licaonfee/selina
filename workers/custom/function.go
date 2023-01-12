// Package custom implements an user defined function
package custom

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/licaonfee/selina"
)

var _ selina.Worker = (*Function)(nil)

// UserFunction define an user custom modification
// is safe to return input to avoid allocations
// if an error is returned Process is aborted
// a filter can be implemented returning (nil,nil)
type UserFunction func(input []byte) ([]byte, error)

// FunctionOptions customize a Function Worker
type FunctionOptions struct {
	Func UserFunction
}

// Check if a combination of options is valid
func (o FunctionOptions) Check() error {
	if o.Func == nil {
		return ErrNilFunction
	}
	return nil
}

// Function allow users to create custom Workers just with a function
type Function struct {
	opts FunctionOptions
}

// ErrNilFunction a nil UserFunction is provided via FunctionOptions
var ErrNilFunction = errors.New("nil UserFunction passed to Worker")

// Process implements selina.Worker interface
func (f *Function) Process(ctx context.Context, args selina.ProcessArgs) error {
	defer close(args.Output)
	if err := f.opts.Check(); err != nil {
		return err
	}
	data := bytes.NewBuffer(nil)
	for {
		select {
		case msg, ok := <-args.Input:
			if !ok {
				return nil
			}
			data.Reset()
			_, _ = io.Copy(data, msg)
			omsg, err := f.opts.Func(data.Bytes())
			if err != nil {
				selina.FreeBuffer(msg)
				return err
			}
			if omsg == nil {
				selina.FreeBuffer(msg)
				continue
			}
			msg.Reset()
			msg.Write(omsg)
			if err := selina.SendContext(ctx, msg, args.Output); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// NewFunction create a Function object with goven options
func NewFunction(opts FunctionOptions) *Function {
	return &Function{opts: opts}
}
