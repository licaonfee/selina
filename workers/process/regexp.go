package process

import (
	"context"
	"regexp"

	"github.com/licaonfee/selina"
)

var _ selina.Worker = (*RegexFilter)(nil)

type RegexFilterOptions struct {
	Pattern string
}

type RegexFilter struct {
	opts RegexFilterOptions
}

func (r *RegexFilter) Process(ctx context.Context, in <-chan []byte, out chan<- []byte) error {
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

func NewRegexpFilter(opts RegexFilterOptions) *RegexFilter {
	return &RegexFilter{opts: opts}
}
