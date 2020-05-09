package ops

import (
	"context"
	"sync"

	"github.com/licaonfee/selina"
	"github.com/robfig/cron/v3"
)

var _ selina.Worker = (*Cron)(nil)

type CronOptions struct {
	Spec    string
	Message []byte
}

type Cron struct {
	opts CronOptions
}

var globalCron *cron.Cron
var initCron sync.Once

func createCron() {
	globalCron = cron.New(cron.WithSeconds())
	globalCron.Start()
}

func (c *Cron) Process(ctx context.Context, args selina.ProcessArgs) error {
	initCron.Do(createCron)
	tick := make(chan struct{})
	bye := make(chan struct{})
	id, err := globalCron.AddFunc(c.opts.Spec, func() {
		select {
		case tick <- struct{}{}:
		case <-bye:
		}
	})
	defer func() {
		globalCron.Remove(id)
		close(bye)
		close(args.Output)
	}()
	if err != nil {
		return err
	}
	for {
		select {
		case _, ok := <-args.Input:
			if !ok {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		case <-tick:
			args.Output <- c.opts.Message
		}
	}
}

func NewCron(opts CronOptions) *Cron {
	return &Cron{opts: opts}
}
