package regex

import (
	"context"
	"regexp"

	"github.com/licaonfee/selina"
)

var _ selina.Worker = (*Filter)(nil)

//FilterOptions customize Filter Worker
type FilterOptions struct {
	//Pattern valid regular expresion
	Pattern string
}

//Check if a combination of options is valid
func (o FilterOptions) Check() error {
	_, err := regexp.Compile(o.Pattern)
	return err
}

//Filter Worker read []byte from input channel, apply regular expresion
// if []byte match against Pattern , []byte is sent to output
type Filter struct {
	opts FilterOptions
}

//Process implements Worker interface
func (r *Filter) Process(ctx context.Context, args selina.ProcessArgs) error {
	if err := r.opts.Check(); err != nil {
		return err
	}
	re, _ := regexp.Compile(r.opts.Pattern)
	defer func() {
		close(args.Output)
	}()
	for {
		select {
		case msg, ok := <-args.Input:
			if ok {
				if re.Match(msg) {
					args.Output <- msg
				}
			} else {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

//NewFilter create a new Filter Worker with specified options
func NewFilter(opts FilterOptions) *Filter {
	return &Filter{opts: opts}
}
