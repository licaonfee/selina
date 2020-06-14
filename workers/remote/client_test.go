package remote_test

import (
	context "context"
	"errors"
	"testing"

	"github.com/licaonfee/selina"
	"github.com/licaonfee/selina/workers"
	"github.com/licaonfee/selina/workers/remote"
)

func TestClientProcess(t *testing.T) {
	tests := []struct {
		name    string
		opts    remote.ClientOptions
		wantErr error
	}{
		//Add test cases
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := remote.NewClient(tt.opts)
			args := selina.ProcessArgs{
				Input:  make(chan []byte),
				Output: make(chan []byte),
			}
			err := c.Process(context.Background(), args)
			if err != tt.wantErr && (!errors.As(err, &tt.wantErr)) {
				t.Errorf("Process() err = %T(%v) , want = %T(%v)", err, err, tt.wantErr, tt.wantErr)
			}
		})
	}
}

func TestClientProcessCancelation(t *testing.T) {
	r := remote.NewClient(remote.ClientOptions{Address: "localhost:7777"})
	if err := workers.ATProcessCancel(r); err != nil {
		t.Fatal(err)
	}
}

func TestClientProcessCloseInput(t *testing.T) {
	r := remote.NewClient(remote.ClientOptions{Address: "localhost:7777"})
	if err := workers.ATProcessCloseInput(r); err != nil {
		t.Fatal(err)
	}
}
func TestClientProcessCloseOutput(t *testing.T) {
	r := remote.NewClient(remote.ClientOptions{Address: "localhost:7777"})
	if err := workers.ATProcessCloseOutput(r); err != nil {
		t.Fatal(err)
	}
}
