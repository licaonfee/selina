package random_test

import (
	"context"
	"testing"

	"github.com/licaonfee/selina"

	"github.com/licaonfee/selina/workers"

	"github.com/licaonfee/selina/workers/random"
)

func TestRandomProcesslen(t *testing.T) {
	tests := []struct {
		name string
		opts random.Options
	}{
		{
			name: "Short slice",
			opts: random.Options{Len: 32},
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
			args := selina.ProcessArgs{Input: input, Output: output}
			if err := r.Process(context.Background(), args); err != nil {
				t.Fatalf("Process() err = %v", err)
				return
			}
			if len(msg) != tt.opts.Len {
				t.Fatalf("Process() got len=%d, want len=%d", len(msg), tt.opts.Len)
			}
		})
	}
}

func TestRandomProcessCancel(t *testing.T) {
	p := random.NewRandom(random.Options{Len: 8})
	if err := workers.ATProcessCancel(p); err != nil {
		t.Fatal(err)
	}
}

func TestRandomProcessCloseInput(t *testing.T) {
	p := random.NewRandom(random.Options{Len: 8})
	if err := workers.ATProcessCloseInput(p); err != nil {
		t.Fatal(err)
	}
}

func TestRandomProcessCloseOutput(t *testing.T) {
	p := random.NewRandom(random.Options{Len: 8})
	if err := workers.ATProcessCloseOutput(p); err != nil {
		t.Fatal(err)
	}
}
