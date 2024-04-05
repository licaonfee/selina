package workers_test

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/licaonfee/selina"
	"github.com/licaonfee/selina/workers"
)

func pass(in []byte) ([]byte, error) {
	return in, nil
}

func TestFunctionProcessCloseInput(t *testing.T) {
	f := workers.NewFunction(workers.FunctionOptions{Func: pass})
	if err := workers.ATProcessCloseInput(f); err != nil {
		t.Fatal(err)
	}
}

func TestFunctionProcessCloseOutput(t *testing.T) {
	f := workers.NewFunction(workers.FunctionOptions{Func: pass})
	if err := workers.ATProcessCloseOutput(f); err != nil {
		t.Fatal(err)
	}
}

func TestFunctionProcessCancel(t *testing.T) {
	f := workers.NewFunction(workers.FunctionOptions{Func: pass})
	if err := workers.ATProcessCancel(f); err != nil {
		t.Fatal(err)
	}
}

func TestFunctionProcess(t *testing.T) {
	tests := []struct {
		name    string
		opts    workers.FunctionOptions
		msgs    []string
		want    []string
		wantErr bool
	}{
		{
			name: "Passthrough",
			opts: workers.FunctionOptions{Func: func(in []byte) ([]byte, error) {
				return in, nil
			}},
			msgs:    []string{"a", "d", "f"},
			want:    []string{"a", "d", "f"},
			wantErr: false,
		},
		{
			name:    "Nil function",
			opts:    workers.FunctionOptions{},
			msgs:    []string{},
			want:    []string{},
			wantErr: true,
		},
		{
			name: "Function error",
			opts: workers.FunctionOptions{Func: func(in []byte) ([]byte, error) {
				return nil, errors.New("function error")
			}},
			msgs:    []string{"une", "dois"},
			want:    []string{},
			wantErr: true,
		},
		{
			name: "Filter function",
			opts: workers.FunctionOptions{Func: func(in []byte) ([]byte, error) {
				if len(in) > 5 {
					return nil, nil
				}
				return in, nil
			}},
			msgs:    []string{"1234", "12345678", "12", "12356789"},
			want:    []string{"1234", "12"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := workers.NewFunction(tt.opts)
			input := selina.SliceAsChannelOfBuffer(tt.msgs, true)
			output := make(chan *bytes.Buffer, len(tt.want))
			args := selina.ProcessArgs{Input: input, Output: output}
			if err := f.Process(context.Background(), args); (err != nil) != tt.wantErr {
				t.Fatalf("Process() unexpected err = %v", err)
			}
			got := []string{}
			for _, b := range selina.ChannelAsSlice(output) {
				got = append(got, b.String())
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Process() got = %#v , want = %#v", got, tt.want)
			}
		})
	}
}
