package remote

import (
	"context"
	"errors"
	"net"
	"sync"

	"github.com/licaonfee/selina"
	"google.golang.org/grpc"
)

var _ WorkerServer = (*Server)(nil)
var _ selina.Worker = (*Server)(nil)
var (
	// ErrDiscarded is returned when there is no slots
	// in process stream
	ErrDiscarded = errors.New("discarded")
)

// ServerOptions customize Server Worker
type ServerOptions struct {
	Network    string
	Address    string
	BufferSize int
}

// Server receive data from a remote endpoint
type Server struct {
	UnimplementedWorkerServer
	opts  ServerOptions
	dataC chan []byte
}

// Send implements grpc service
func (s *Server) Send(ctx context.Context, msg *Message) (*Error, error) {
	select {
	case <-ctx.Done():
		return &Error{Message: ctx.Err().Error()}, ctx.Err()
	case s.dataC <- msg.Data:
		return &Error{}, nil
	}
}

// Push put a []byte into process stream, return ErrDiscarded if
// msg is not send immediately
func (s *Server) Push(msg []byte) error {
	select {
	case s.dataC <- msg:
		return nil
	default:
		return ErrDiscarded
	}
}

// Process implements selina.Worker interface
func (s *Server) Process(ctx context.Context, args selina.ProcessArgs) (errp error) {
	defer close(args.Output)
	wg := sync.WaitGroup{}
	wg.Add(1)
	listener, err := net.Listen(s.opts.Network, s.opts.Address)
	if err != nil {
		return err
	}

	gserver := grpc.NewServer()
	RegisterWorkerServer(gserver, s)
	var cerr error
	defer func() {
		wg.Wait()
		if errp == nil && cerr != nil {
			errp = cerr
		}
	}()
	go func() {
		cerr = gserver.Serve(listener)
		wg.Done()
	}()
	defer gserver.GracefulStop()
	for {
		select {
		case _, ok := <-args.Input:
			if !ok {
				return nil
			}
		case msg := <-s.dataC:
			if err := selina.SendContext(ctx, msg, args.Output); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()

		}
	}
}

// NewServer create a new grpc server with given options
func NewServer(opts ServerOptions) *Server {
	return &Server{opts: opts,
		dataC: make(chan []byte, opts.BufferSize)}
}
