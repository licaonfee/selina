//+build !unit

package remote_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/licaonfee/selina"
	"github.com/licaonfee/selina/workers/remote"
	"github.com/licaonfee/selina/workers/text"
)

const messageLength = 16
const address = "localhost:65000"

func producer(data string) selina.Pipeliner {
	txt := text.NewReader(text.ReaderOptions{Reader: strings.NewReader(data)})
	client := remote.NewClient(remote.ClientOptions{Address: address})
	return selina.LinealPipeline(
		selina.NewNode("input", txt),
		selina.NewNode("grpc", client))
}

func consumer(w io.Writer) selina.Pipeliner {
	server := remote.NewServer(remote.ServerOptions{Network: "tcp", Address: address})
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
			ctx, cancel := context.WithCancel(context.Background())
			go c.Run(ctx)
			time.Sleep(time.Second)
			go p.Run(ctx)
			time.Sleep(time.Second)
			cancel()
			time.Sleep(time.Second)
			got := recv.String()
			if tt.data != got {
				t.Errorf("Data transfer failed: got = '%v' , want = '%v'", got, tt.data)
			}
		})
	}
}
