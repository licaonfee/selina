package regex_test

import (
	"context"
	"reflect"

	"github.com/licaonfee/selina/workers"
	"github.com/licaonfee/selina/workers/regex"

	"github.com/licaonfee/selina"

	"testing"
)

func TestRegexFilter_Process(t *testing.T) {
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
			r := regex.NewRegexpFilter(tt.args.opts)
			input := selina.SliceAsChannel(tt.args.in, true)
			output := make(chan []byte)
			got := []string{}
			wait := make(chan struct{})
			go func() {
				got = selina.ChannelAsSlice(output)
				close(wait)
			}()

			if err := r.Process(context.Background(), input, output); (err != nil) != tt.wantErr {
				t.Fatalf("RegexFilter.Process() error = %v", err)
			}
			if tt.wantErr {
				return
			}
			<-wait
			if !reflect.DeepEqual(got, tt.args.want) {
				t.Fatalf("Process() got = %T, want =  %v", got, tt.args.want)
			}
		})
	}
}

func TestRegexFilter_Process_cancelation(t *testing.T) {
	r := regex.NewRegexpFilter(regex.FilterOptions{Pattern: ".*"})
	workers.ATProcessCancel(r, t)
}

func TestRegexFilter_Process_close_input(t *testing.T) {
	r := regex.NewRegexpFilter(regex.FilterOptions{Pattern: ".*"})
	workers.ATProcessCloseInput(r, t)
}
func TestRegexFilter_Process_close_output(t *testing.T) {
	r := regex.NewRegexpFilter(regex.FilterOptions{Pattern: ".*"})
	workers.ATProcessCloseOutput(r, t)
}
