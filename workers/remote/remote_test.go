//+build !unit

package remote_test

import (
	"bytes"
	"context"
	"io"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/licaonfee/selina"
	"github.com/licaonfee/selina/workers/remote"
	"github.com/licaonfee/selina/workers/text"
	"golang.org/x/sync/errgroup"
)

const address = "localhost:"
const baseport = 65000

var autoport int = 0

func producer(data string) selina.Pipeliner {
	txt := text.NewReader(text.ReaderOptions{Reader: strings.NewReader(data)})
	autoport++
	client := remote.NewClient(remote.ClientOptions{Address: address + strconv.Itoa(baseport+autoport)})
	return selina.LinealPipeline(
		selina.NewNode("input", txt),
		selina.NewNode("grpc", client))
}

func consumer(w io.Writer) selina.Pipeliner {
	server := remote.NewServer(remote.ServerOptions{Network: "tcp", Address: address + strconv.Itoa(baseport+autoport)})
	txt := text.NewWriter(text.WriterOptions{Writer: w})
	return selina.LinealPipeline(
		selina.NewNode("grpc", server),
		selina.NewNode("output", txt))
}

func TestRemote(t *testing.T) {
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
			eg, ctx := errgroup.WithContext(context.Background())
			consCtx, cancel := context.WithTimeout(ctx, time.Second*5)
			eg.Go(func() error {
				return c.Run(consCtx)
			})
			eg.Go(func() error {
				err := p.Run(ctx)
				return err
			})
			if err := eg.Wait(); err != nil && err != context.DeadlineExceeded {
				t.Errorf("unexpected error = %v", err)
			}
			cancel()
			got := recv.String()
			if tt.data != got {
				t.Errorf("Data transfer failed: got = '%v' , want = '%v'", got, tt.data)
			}
		})
	}
}
