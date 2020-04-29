package random_test

import (
	"context"
	"testing"

	"github.com/licaonfee/selina/workers"

	"github.com/licaonfee/selina/workers/random"
)

func TestRandom_Process_len(t *testing.T) {
	tests := []struct {
		name string
		opts random.RandomOptions
	}{
		{
			name: "Short slice",
			opts: random.RandomOptions{Len: 32},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := random.NewRandom(tt.opts)
			input := make(chan []byte)
			output := make(chan []byte)
			var msg []byte
			go func() {
				msg = <-output
				close(input)
				for range output {
				}
			}()
			if err := r.Process(context.Background(), input, output); err != nil {
				t.Fatalf("Process() err = %v", err)
				return
			}
			if len(msg) != tt.opts.Len {
				t.Fatalf("Process() got len=%d, want len=%d", len(msg), tt.opts.Len)
			}
		})
	}
}

func TestRandom_Process_cancel(t *testing.T) {
	p := random.NewRandom(random.RandomOptions{Len: 8})
	workers.ATProcessCancel(p, t)
}

func TestRandom_Process_close_input(t *testing.T) {
	p := random.NewRandom(random.RandomOptions{Len: 8})
	workers.ATProcessCloseInput(p, t)
}

func TestRandom_Process_close_output(t *testing.T) {
	p := random.NewRandom(random.RandomOptions{Len: 8})
	workers.ATProcessCloseOutput(p, t)
}
