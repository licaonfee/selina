package workers_test

import (
	"context"
	"testing"

	"github.com/licaonfee/selina/workers"

	"github.com/licaonfee/selina"
)

//this worker implements all required functionality
type idealWorker struct{}

func (i *idealWorker) Process(ctx context.Context, input <-chan []byte, output chan<- []byte) error {
	//A Worker must close output in all exit paths
	defer close(output)
	//when input is closed return nil
	for {
		select {
		case _, ok := <-input:
			if !ok { //when input is closed and drained return nil
				return nil
			}
		case <-ctx.Done():
			return ctx.Err() //On cancel return context.Canceled
		}
	}
}

type badContextWorker struct{}

func (b *badContextWorker) Process(ctx context.Context, input <-chan []byte, output chan<- []byte) error {
	defer close(output)
	//just ignore context
	for range input {
	}
	return nil
}

type badInputWorker struct{}

func (b *badInputWorker) Process(ctx context.Context, input <-chan []byte, output chan<- []byte) error {
	defer close(output)
	//ignore input channel
	<-ctx.Done()
	return ctx.Err()
}

type badOutputWorker struct{}

func (b *badOutputWorker) Process(ctx context.Context, input <-chan []byte, output chan<- []byte) error {
	for {
		select {
		case _, ok := <-input:
			if !ok { //when input is closed and drained return nil
				return nil
			}
		case <-ctx.Done():
			return ctx.Err() //On cancel return context.Canceled
		}
	}
}

//Test if all Acceptance tests pass and fail in correct cases
func TestAcceptanceTests(t *testing.T) {
	tests := []struct {
		name    string
		at      func(selina.Worker) error
		w       selina.Worker
		wantErr bool
	}{
		{
			name:    "Cancel Process OK",
			at:      workers.ATProcessCancel,
			w:       &idealWorker{},
			wantErr: false,
		},
		{
			name:    "Close input OK",
			at:      workers.ATProcessCloseInput,
			w:       &idealWorker{},
			wantErr: false,
		},
		{
			name:    "Close output OK",
			at:      workers.ATProcessCloseOutput,
			w:       &idealWorker{},
			wantErr: false,
		},
		{
			name:    "Cancel ignore context",
			at:      workers.ATProcessCancel,
			w:       &badContextWorker{},
			wantErr: true,
		},
		{
			name:    "Close input is ignored",
			at:      workers.ATProcessCloseInput,
			w:       &badInputWorker{},
			wantErr: true,
		},
		{
			name:    "output is not closed",
			at:      workers.ATProcessCloseOutput,
			w:       &badOutputWorker{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.at(tt.w); (err != nil) != tt.wantErr {
				t.Fatalf("Acceptance test is broken: got unexpected err = %v", err)
			}
		})
	}
}
