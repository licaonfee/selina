package ops_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/licaonfee/selina"
	"golang.org/x/net/context"

	"github.com/licaonfee/selina/workers"
	"github.com/licaonfee/selina/workers/ops"
)

func TestCronProcessCancelation(t *testing.T) {
	c := ops.NewCron(ops.CronOptions{Spec: "@every 10s"})
	if err := workers.ATProcessCancel(c); err != nil {
		t.Fatal(err)
	}
}

func TestCronProcessCloseInput(t *testing.T) {
	c := ops.NewCron(ops.CronOptions{Spec: "@every 1s"})
	if err := workers.ATProcessCloseInput(c); err != nil {
		t.Fatal(err)
	}
}
func TestCronProcessCloseOutput(t *testing.T) {
	c := ops.NewCron(ops.CronOptions{Spec: "@every 1s"})
	if err := workers.ATProcessCloseOutput(c); err != nil {
		t.Fatal(err)
	}
}

func TestCronProcess(t *testing.T) {
	tests := []struct {
		name    string
		opts    ops.CronOptions
		want    []string
		runFor  time.Duration
		wantErr error
	}{
		{
			name:    "Invalid spec",
			opts:    ops.CronOptions{Spec: ""},
			want:    []string{},
			wantErr: ops.ErrBadCronSpec,
		},
		{
			name:    "Tick message",
			opts:    ops.CronOptions{Spec: "@every 1s", Message: []byte("foo")},
			want:    []string{"foo"},
			runFor:  time.Second,
			wantErr: nil,
		},
		{
			name:    "Tick nil",
			opts:    ops.CronOptions{Spec: "@every 1s"},
			want:    []string{""},
			runFor:  time.Second,
			wantErr: nil,
		},
		{
			name:    "Bad spec",
			opts:    ops.CronOptions{Spec: "@eberi 1z"},
			want:    []string{},
			runFor:  time.Second,
			wantErr: ops.ErrBadCronSpec,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := ops.NewCron(tt.opts)
			input := make(chan []byte)
			output := make(chan []byte, len(tt.want))
			args := selina.ProcessArgs{Input: input, Output: output}
			go func(wait time.Duration) {
				time.Sleep(wait)
				close(input)
			}(tt.runFor)
			if err := c.Process(context.Background(), args); !errors.Is(err, tt.wantErr) {
				t.Fatalf("Process() err = %v, wantErr %v", err, tt.wantErr)
			}
			got := selina.ChannelAsSlice(output)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Process() got = %#v , want = %#v", got, tt.want)
			}

		})
	}
}
