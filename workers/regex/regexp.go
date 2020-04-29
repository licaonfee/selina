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

//Filter Worker read []byte from input channel, apply regular expresion
// if []byte match against Pattern , []byte is sent to output
type Filter struct {
	opts FilterOptions
}

//Process implements Worker interface
func (r *Filter) Process(ctx context.Context, in <-chan []byte, out chan<- []byte) error {
	re, err := regexp.Compile(r.opts.Pattern)
	if err != nil {
		return err
	}
	defer func() {
		close(out)
	}()
	for {
		select {
		case msg, ok := <-in:
			if ok {
				if re.Match(msg) {
					out <- msg
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
