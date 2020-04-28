package workers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/licaonfee/selina"
)

const waitProcessSleepDuration = time.Millisecond * 50
const closeInputTimeout = time.Millisecond * 50

//ATPProcessCancel a worker must terminate and return context.Canceled
// when context is canceled
func ATProcessCancel(w selina.Worker, t *testing.T) {
	input := make(chan []byte)
	output := make(chan []byte) //unbuffered so, process wait forever
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(waitProcessSleepDuration) //wait to start process
		cancel()
	}()
	if err := w.Process(ctx, input, output); err != context.Canceled {
		t.Fatalf("Process() err = %v", err)
	}
}

//ATPProcessCloseInput a worker must terminate and return nil
// when input chanel (<-chan []byte )is closed
func ATProcessCloseInput(w selina.Worker, t *testing.T) {
	input := make(chan []byte)
	output := make(chan []byte)
	resp := make(chan error, 1)
	go func() {
		resp <- w.Process(context.Background(), input, output)
	}()
	go func() {
		for range output {
			//Consume output to avoid Process lock
		}
	}()
	time.Sleep(waitProcessSleepDuration)
	close(input)
	select {
	case err := <-resp:
		if err != nil {
			t.Fatalf("Process() err = %v , want = nil", err)
		}
	case <-time.After(closeInputTimeout):
		t.Fatalf("Process() does not terminate on closed input")
	}
}

//ATPProcessCloseOutput a worker must close its output channel on exit
func ATProcessCloseOutput(w selina.Worker, t *testing.T) {
	input := make(chan []byte)
	output := make(chan []byte)
	close(input)
	if err := w.Process(context.Background(), input, output); err != nil {
		t.Fatalf("Process() err = %v", err)
	}
	go func() {
		for range output {
		}
	}()
	errC := make(chan error)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errC <- nil
			}
		}()
		close(output)
		errC <- errors.New("channel is not closed")
	}()
	if err := <-errC; err != nil {
		t.Fatalf("Process() err = %v", err)
	}
}
