package remote

import (
	"context"

	"github.com/licaonfee/selina"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var _ selina.Worker = (*Client)(nil)

// ClientOptions customize client
type ClientOptions struct {
	Address string
}

// Client connect to a remote grpc endpoint
type Client struct {
	opts ClientOptions
}

// Process implements selina.Worker interface
func (c *Client) Process(ctx context.Context, args selina.ProcessArgs) error {
	defer close(args.Output)
	conn, err := grpc.DialContext(ctx, c.opts.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	//TODO: handle error
	defer conn.Close()
	wc := NewWorkerClient(conn)
	for {
		select {
		case msg, ok := <-args.Input:
			if !ok {
				return nil
			}
			data := make([]byte, len(msg.Bytes()))
			copy(data, msg.Bytes())
			selina.FreeBuffer(msg)
			m := Message{Data: data}
			_, err := wc.Send(ctx, &m)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// NewClient create a new Client with given options
func NewClient(opts ClientOptions) *Client {
	return &Client{opts: opts}
}
