package workers

import (
	"context"
	"errors"
	"time"

	"github.com/licaonfee/selina"
)

const waitProcessSleepDuration = time.Millisecond * 50
const closeInputTimeout = time.Millisecond * 50

//ErrProcessIgnoreCtx worker.Process does not terminate when context is canceled
var ErrProcessIgnoreCtx = errors.New("ignored context.Done")

//ErrNotTerminatedOnCloseInput worker.Process does not finish when input channel is closed
var ErrNotTerminatedOnCloseInput = errors.New("not terminate on closed input")

//ErrOutputNotClosed worker.Process does not close output channel
var ErrOutputNotClosed = errors.New("output channel is not closed")

//ATProcessCancel a worker must terminate and return context.Canceled
// when context is canceled
func ATProcessCancel(w selina.Worker) error {
	input := make(chan []byte)
	output := make(chan []byte) //unbuffered so, process wait forever
	ctx, cancel := context.WithCancel(context.Background())
	errC := make(chan error, 1)
	go func() {
		time.Sleep(waitProcessSleepDuration) //wait to start process
		cancel()
	}()
	go func() {
		args := selina.ProcessArgs{Input: input, Output: output}
		errC <- w.Process(ctx, args)
	}()
	<-ctx.Done() //wait until context is canceled
	select {
	case err := <-errC:
		if err != context.Canceled {
			return err
		}
		return nil
	case <-time.After(waitProcessSleepDuration):
		return ErrProcessIgnoreCtx
	}

}

//ATProcessCloseInput a worker must finish its job and return nil
// when input chanel (<-chan []byte )is closed
func ATProcessCloseInput(w selina.Worker) error {
	input := make(chan []byte)
	output := make(chan []byte)
	resp := make(chan error, 1)
	go func() {
		args := selina.ProcessArgs{Input: input, Output: output}
		resp <- w.Process(context.Background(), args)
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
		return err
	case <-time.After(closeInputTimeout):
		return ErrNotTerminatedOnCloseInput
	}
}

//ATProcessCloseOutput a worker must close its output channel on exit
func ATProcessCloseOutput(w selina.Worker) error {
	input := make(chan []byte)
	output := make(chan []byte)
	close(input)
	args := selina.ProcessArgs{Input: input, Output: output}
	_ = w.Process(context.Background(), args)
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
		errC <- ErrOutputNotClosed
	}()
	if err := <-errC; err != nil {
		return err
	}
	return nil
}
