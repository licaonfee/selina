package workers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/licaonfee/selina"
)

const waitProcessSleepDuration = time.Millisecond * 50
const closeInputTimeout = time.Millisecond * 50

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
			return fmt.Errorf("Process() err = %v", err)
		}
		return nil
	case <-time.After(waitProcessSleepDuration):
		return fmt.Errorf("Process() ignore ctx.Done()")
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
		return fmt.Errorf("Process() does not terminate on closed input")
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
		errC <- errors.New("channel is not closed")
	}()
	if err := <-errC; err != nil {
		return fmt.Errorf("Process() err = %v", err)
	}
	return nil
}
