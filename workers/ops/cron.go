package ops

import (
	"context"
	"sync"

	"github.com/licaonfee/selina"
	"github.com/robfig/cron/v3"
)

var _ selina.Worker = (*Cron)(nil)

const cronDefaultOptions = cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor

//CronOptions customize cron behaviour
type CronOptions struct {
	//Spec use same format as github.com/robfig/cron/v3
	//Second Minute Hour DayOfMonth Month DayOfWeek
	Spec string
	//Which message will be sent every schedule
	Message []byte
}

//Check if a combination of options is valid
func (o CronOptions) Check() error {
	p := cron.NewParser(cronDefaultOptions)
	_, err := p.Parse(o.Spec)
	return err
}

//Cron send an specifi message at scheduled intervals
type Cron struct {
	opts CronOptions
}

var globalCron *cron.Cron
var initCron sync.Once

func createCron() {
	p := cron.NewParser(cronDefaultOptions)
	globalCron = cron.New(cron.WithParser(p))
	globalCron.Start()
}

//Process add a job scec, any message received will be discarded
// when input is closed this worker return nil
func (c *Cron) Process(ctx context.Context, args selina.ProcessArgs) error {
	defer close(args.Output)
	if err := c.opts.Check(); err != nil {
		return err
	}
	initCron.Do(createCron)
	tick := make(chan struct{})
	bye := make(chan struct{})
	id, _ := globalCron.AddFunc(c.opts.Spec, func() {
		select {
		case tick <- struct{}{}:
		case <-bye:
		}
	})
	defer func() {
		globalCron.Remove(id)
		close(bye)
	}()
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

//NewCron create a new Cron Worker with given options
func NewCron(opts CronOptions) *Cron {
	return &Cron{opts: opts}
}
