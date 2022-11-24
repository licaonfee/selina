package ops

import (
	"context"
	"time"

	"github.com/licaonfee/selina"
	"github.com/licaonfee/tserie"
)

var _ selina.Worker = (*TimeSerie)(nil)

// TimeSerieOptions customize optiosn for generation
type TimeSerieOptions struct {
	Start       time.Time
	Stop        time.Time
	Step        time.Duration
	Generator   func(time.Time) float64
	WriteFormat selina.Marshaler
}

// TimeSerie generate time and values in a given time range
type TimeSerie struct {
	opts TimeSerieOptions
}

// Process generate timeseries and put it out in a channel
func (t *TimeSerie) Process(ctx context.Context, args selina.ProcessArgs) error {
	defer close(args.Output)
	ts := tserie.NewTimeIterator(t.opts.Start, t.opts.Stop, t.opts.Step, t.opts.Generator)
	if t.opts.WriteFormat == nil {
		t.opts.WriteFormat = selina.DefaultMarshaler
	}
	for {
		select {
		case _, ok := <-args.Input:
			if !ok {
				return nil
			}
		default:
			if !ts.Next() {
				return nil
			}
			b, err := t.opts.WriteFormat(ts.Item())
			if err != nil {
				return err
			}
			if err := selina.SendContext(ctx, b, args.Output); err != nil {
				return err
			}
		}
	}
}

// NewTimeSerie create a new TimeSerie generator with given options
func NewTimeSerie(opts TimeSerieOptions) *TimeSerie {
	return &TimeSerie{opts: opts}
}
