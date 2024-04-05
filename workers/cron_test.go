package workers_test

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/licaonfee/selina"
	"golang.org/x/net/context"

	"github.com/licaonfee/selina/workers"
)

func TestCronProcessCancelation(t *testing.T) {
	c := workers.NewCron(workers.CronOptions{Spec: "@every 10s"})
	if err := workers.ATProcessCancel(c); err != nil {
		t.Fatal(err)
	}
}

func TestCronProcessCloseInput(t *testing.T) {
	c := workers.NewCron(workers.CronOptions{Spec: "@every 1s"})
	if err := workers.ATProcessCloseInput(c); err != nil {
		t.Fatal(err)
	}
}
func TestCronProcessCloseOutput(t *testing.T) {
	c := workers.NewCron(workers.CronOptions{Spec: "@every 1s"})
	if err := workers.ATProcessCloseOutput(c); err != nil {
		t.Fatal(err)
	}
}

func TestCronProcess(t *testing.T) {
	tests := []struct {
		name    string
		opts    workers.CronOptions
		want    []string
		runFor  time.Duration
		wantErr error
	}{
		{
			name:    "Invalid spec",
			opts:    workers.CronOptions{Spec: ""},
			want:    []string{},
			wantErr: workers.ErrBadCronSpec,
		},
		{
			name:    "Tick message",
			opts:    workers.CronOptions{Spec: "@every 1s", Message: []byte("foo")},
			want:    []string{"foo"},
			runFor:  time.Second,
			wantErr: nil,
		},
		{
			name:    "Tick nil",
			opts:    workers.CronOptions{Spec: "@every 1s"},
			want:    []string{""},
			runFor:  time.Second,
			wantErr: nil,
		},
		{
			name:    "Bad spec",
			opts:    workers.CronOptions{Spec: "@eberi 1z"},
			want:    []string{},
			runFor:  time.Second,
			wantErr: workers.ErrBadCronSpec,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := workers.NewCron(tt.opts)
			input := make(chan *bytes.Buffer)
			output := make(chan *bytes.Buffer, len(tt.want))
			args := selina.ProcessArgs{Input: input, Output: output}
			go func(wait time.Duration) {
				time.Sleep(wait)
				close(input)
			}(tt.runFor)
			if err := c.Process(context.Background(), args); !errors.Is(err, tt.wantErr) {
				t.Fatalf("Process() err = %v, wantErr %v", err, tt.wantErr)
			}
			got := []string{}
			for _, b := range selina.ChannelAsSlice(output) {
				got = append(got, b.String())
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Process() got = %#v , want = %#v", got, tt.want)
			}

		})
	}
}
