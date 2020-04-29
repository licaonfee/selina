package selina

import "context"

//Worker is standard interface implemented by proccessors, is used to build pipeline nodes
// All Worker implementations must meet the following conditions
// On close input channel, Process must finalize its work gracefully, and return nil
// On context cancellation, Process finalize ASAP and return context.Cancelled
// On finish, Process must close output channel and return error or nil
type Worker interface {
	//Process must close write only channel
	Process(ctx context.Context, input <-chan []byte, output chan<- []byte) error
}
