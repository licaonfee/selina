package remote_test

import (
	"context"
	"errors"
	"net"
	"reflect"
	"testing"

	"github.com/licaonfee/selina"
	"github.com/licaonfee/selina/workers"
	"github.com/licaonfee/selina/workers/remote"
)

func TestServerProcess(t *testing.T) {
	tests := []struct {
		name string
		//BufferSize must be equal to len(send)
		opts remote.ServerOptions
		//for empty use [][]byte{}
		send    [][]byte
		wantErr any
	}{
		{
			name:    "Success no messages",
			opts:    remote.ServerOptions{Network: "tcp", Address: ":0"},
			send:    [][]byte{},
			wantErr: nil,
		},
		{
			name:    "Success",
			opts:    remote.ServerOptions{Network: "tcp", Address: ":0", BufferSize: 2},
			send:    [][]byte{[]byte("foo"), []byte("bar")},
			wantErr: nil,
		},
		{
			name:    "Bad Options",
			opts:    remote.ServerOptions{Network: "tkp", Address: ":0"},
			send:    [][]byte{},
			wantErr: &net.OpError{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := remote.NewServer(tt.opts)
			input := make(chan []byte)
			output := make(chan []byte, len(tt.send))
			args := selina.ProcessArgs{
				Input:  input,
				Output: output}

			for i := 0; i < len(tt.send); i++ {
				_ = s.Push(tt.send[i])
			}
			ec := make(chan error)
			go func() {
				ec <- s.Process(context.Background(), args)
			}()
			got := make([][]byte, 0)
			for i := 0; i < len(tt.send); i++ {
				msg := <-output
				got = append(got, msg)
			}
			close(input)
			if err := <-ec; (err != tt.wantErr) && (!errors.As(err, &tt.wantErr)) {
				t.Errorf("Process() err = %T(%v), want = %T(%v)", err, err, tt.wantErr, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.send) {
				t.Errorf("Process() got = %#v, want = %#v", got, tt.send)
			}
		})
	}
}

func TestServerProcessCancelation(t *testing.T) {
	r := remote.NewServer(remote.ServerOptions{Network: "tcp", Address: ":0"})
	if err := workers.ATProcessCancel(r); err != nil {
		t.Fatal(err)
	}
}

func TestServerProcessCloseInput(t *testing.T) {
	r := remote.NewServer(remote.ServerOptions{Network: "tcp", Address: ":0"})
	if err := workers.ATProcessCloseInput(r); err != nil {
		t.Fatal(err)
	}
}
func TestServerProcessCloseOutput(t *testing.T) {
	r := remote.NewServer(remote.ServerOptions{Network: "tcp", Address: ":0"})
	if err := workers.ATProcessCloseOutput(r); err != nil {
		t.Fatal(err)
	}
}
