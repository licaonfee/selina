package workers_test

import (
	"context"
	"errors"
	"testing"

	"github.com/licaonfee/selina/workers"

	"github.com/licaonfee/selina"
)

var _ selina.Worker = (*idealWorker)(nil)

//this worker implements all required functionality
type idealWorker struct{}

func (i *idealWorker) Process(ctx context.Context, args selina.ProcessArgs) error {
	//A Worker must close output in all exit paths
	defer close(args.Output)
	//when input is closed return nil
	for {
		select {
		case _, ok := <-args.Input:
			if !ok { //when input is closed and drained return nil
				return nil
			}
		case <-ctx.Done():
			return ctx.Err() //On cancel return context.Canceled
		}
	}
}

var _ selina.Worker = (*badContextWorker)(nil)

type badContextWorker struct{}

func (b *badContextWorker) Process(ctx context.Context, args selina.ProcessArgs) error {
	defer close(args.Output)
	//just ignore context
	for range args.Input {
	}
	return nil
}

var _ selina.Worker = (*badInputWorker)(nil)

type badInputWorker struct{}

func (b *badInputWorker) Process(ctx context.Context, args selina.ProcessArgs) error {
	defer close(args.Output)
	//ignore input channel
	<-ctx.Done()
	return ctx.Err()
}

var _ selina.Worker = (*badOutputWorker)(nil)

type badOutputWorker struct{}

func (b *badOutputWorker) Process(ctx context.Context, args selina.ProcessArgs) error {
	for {
		select {
		case _, ok := <-args.Input:
			if !ok { //when input is closed and drained return nil
				return nil
			}
		case <-ctx.Done():
			return ctx.Err() //On cancel return context.Canceled
		}
	}
}

type workerN struct {
	N   int
	Msg []byte
}

func (w *workerN) Process(ctx context.Context, args selina.ProcessArgs) error {
	defer close(args.Output)
	for i := 0; i < w.N; i++ {
		select {
		case args.Output <- w.Msg:
		case <-ctx.Done():
			return ctx.Err()
		case _, ok := <-args.Input:
			if !ok {
				return nil
			}
		}
	}
	return nil
}

var _ selina.Worker = (*badCtxPropagation)(nil)

type badCtxPropagation struct{}

var errBadCtxError = errors.New("no context propagation")

func (b *badCtxPropagation) Process(ctx context.Context, args selina.ProcessArgs) error {
	//A Worker must close output in all exit paths
	defer close(args.Output)
	//when input is closed return nil
	for {
		select {
		case _, ok := <-args.Input:
			if !ok { //when input is closed and drained return nil
				return nil
			}
		case <-ctx.Done():
			return errBadCtxError
		}
	}
}

//Test if all Acceptance tests pass and fail in correct cases
func TestAcceptanceTests(t *testing.T) {
	tests := []struct {
		name    string
		at      func(selina.Worker) error
		w       selina.Worker
		wantErr error
	}{
		{
			name:    "Cancel Process OK",
			at:      workers.ATProcessCancel,
			w:       &idealWorker{},
			wantErr: nil,
		},
		{
			name:    "Close input OK",
			at:      workers.ATProcessCloseInput,
			w:       &idealWorker{},
			wantErr: nil,
		},
		{
			name:    "Close output OK",
			at:      workers.ATProcessCloseOutput,
			w:       &idealWorker{},
			wantErr: nil,
		},
		{
			name:    "Cancel ignore context",
			at:      workers.ATProcessCancel,
			w:       &badContextWorker{},
			wantErr: workers.ErrProcessIgnoreCtx,
		},
		{
			name:    "Cancel bad propagation",
			at:      workers.ATProcessCancel,
			w:       &badCtxPropagation{},
			wantErr: errBadCtxError,
		},
		{
			name:    "Close input is ignored",
			at:      workers.ATProcessCloseInput,
			w:       &badInputWorker{},
			wantErr: workers.ErrNotTerminatedOnCloseInput,
		},
		{
			name:    "output is not closed",
			at:      workers.ATProcessCloseOutput,
			w:       &badOutputWorker{},
			wantErr: workers.ErrOutputNotClosed,
		},
		{
			name:    "No lock test on output",
			at:      workers.ATProcessCloseOutput,
			w:       &workerN{N: 10, Msg: []byte("message")},
			wantErr: nil,
		},
		{
			name:    "No lock test on input",
			at:      workers.ATProcessCloseInput,
			w:       &workerN{N: 10, Msg: []byte("message")},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.at(tt.w); err != tt.wantErr {
				t.Fatalf("Acceptance test is broken: got unexpected err = %v", err)
			}
		})
	}
}
