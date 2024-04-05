//go:build !unit
// +build !unit

package remote_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/licaonfee/selina"
	"github.com/licaonfee/selina/workers"
	"github.com/licaonfee/selina/workers/remote"
	"golang.org/x/sync/errgroup"
)

const address = "localhost:65000"

func producer(data string) selina.Pipeliner {
	txt := workers.NewTextReader(workers.TextReaderOptions{Reader: strings.NewReader(data)})
	client := remote.NewClient(remote.ClientOptions{Address: address})
	return selina.LinealPipeline(
		selina.NewNode("input", txt),
		selina.NewNode("grpc", client))
}

func consumer(w io.Writer) selina.Pipeliner {
	server := remote.NewServer(remote.ServerOptions{Network: "tcp", Address: address})
	txt := workers.NewTextWriter(workers.TextWriterOptions{Writer: w})
	return selina.LinealPipeline(
		selina.NewNode("grpc", server),
		selina.NewNode("output", txt))
}

func TestRemote(t *testing.T) {
	const pauseDuration = time.Millisecond * 250
	tests := []struct {
		name string
		data string
	}{
		{
			name: "Single message",
			data: "a single line message\n",
		},
		{
			name: "Multiple messages",
			data: `This is a 
			multiple line message
			spaces and new lines must 
			be present` + "\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var recv bytes.Buffer
			p := producer(tt.data)
			c := consumer(&recv)
			ctxParent, cancel := context.WithCancel(context.Background())
			eg, ctx := errgroup.WithContext(ctxParent)
			eg.Go(func() error {
				return c.Run(ctx)
			})
			time.Sleep(pauseDuration)
			eg.Go(func() error {
				return p.Run(ctx)
			})
			time.Sleep(pauseDuration)
			eg.Go(func() error {
				cancel()
				return nil
			})
			if err := eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
				t.Errorf("unexpected error = %v", err)
			}
			got := recv.String()
			if tt.data != got {
				t.Errorf("Data transfer failed: got = '%v' , want = '%v'", got, tt.data)
			}
		})
	}
}
