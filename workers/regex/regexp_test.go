package regex_test

import (
	"bytes"
	"context"
	"reflect"

	"github.com/licaonfee/selina/workers"
	"github.com/licaonfee/selina/workers/regex"

	"github.com/licaonfee/selina"

	"testing"
)

func TestFilter_Process(t *testing.T) {
	type args struct {
		opts regex.FilterOptions
		in   []string
		want []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Simple FIlter",
			args: args{
				opts: regex.FilterOptions{Pattern: "^ba.+"},
				in:   []string{"foo", "bar", "baz"},
				want: []string{"bar", "baz"},
			},
			wantErr: false,
		},
		{
			name: "Invalid Filter",
			args: args{
				opts: regex.FilterOptions{Pattern: "[----"},
				in:   []string{"you", "shall", "not", "pass"},
				want: []string{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := regex.NewFilter(tt.args.opts)
			input := selina.SliceAsChannelOfBuffer(tt.args.in, true)
			output := make(chan *bytes.Buffer)
			got := []*bytes.Buffer{}
			wait := make(chan struct{})
			go func() {
				got = selina.ChannelAsSlice(output)
				close(wait)
			}()
			args := selina.ProcessArgs{Input: input, Output: output}
			if err := r.Process(context.Background(), args); (err != nil) != tt.wantErr {
				t.Fatalf("RegexFilter.Process() error = %v", err)
			}
			if tt.wantErr {
				return
			}
			<-wait
			realGot := []string{}
			for _, b := range got {
				realGot = append(realGot, b.String())
			}
			if !reflect.DeepEqual(realGot, tt.args.want) {
				t.Fatalf("Process() got = %T, want =  %v", got, tt.args.want)
			}
		})
	}
}

func TestFilterProcessCancelation(t *testing.T) {
	r := regex.NewFilter(regex.FilterOptions{Pattern: ".*"})
	if err := workers.ATProcessCancel(r); err != nil {
		t.Fatal(err)
	}
}

func TestFilterProcessCloseInput(t *testing.T) {
	r := regex.NewFilter(regex.FilterOptions{Pattern: ".*"})
	if err := workers.ATProcessCloseInput(r); err != nil {
		t.Fatal(err)
	}
}
func TestRegexFilterProcessCloseOutput(t *testing.T) {
	r := regex.NewFilter(regex.FilterOptions{Pattern: ".*"})
	if err := workers.ATProcessCloseOutput(r); err != nil {
		t.Fatal(err)
	}
}
