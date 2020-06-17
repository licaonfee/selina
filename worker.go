package selina

import (
	"context"
	"errors"
)

//Worker is standard interface implemented by proccessors, is used to build pipeline nodes
// All Worker implementations must meet the following conditions
// if a worker does not have another worker in upstream then its receive a nil channel in input
// this is useful to idetify the situation and return and error
// On close input channel, Process must finalize its work gracefully, and return nil
// On context cancellation, Process finalize ASAP and return context.Cancelled
// On finish, Process must close output channel and return error or nil
type Worker interface {
	//Process must close write only channel
	Process(ctx context.Context, args ProcessArgs) error
}

//ErrNilUpstream is returned when a worker does not allow to not have an upstream worker
var ErrNilUpstream = errors.New("nil upstream channel")

//ProcessArgs encapsulate arguments to Worker.Process
type ProcessArgs struct {
	//Input is nil when there is no upstream channel
	Input  <-chan []byte
	Output chan<- []byte
	Err    chan error
}

type WrokerOptions interface {
	//Check return an error if options has an invalid value
	//it must not modify values at all and by preference should be
	//implemented as a value receiver
	Check() error
}
