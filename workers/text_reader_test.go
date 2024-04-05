package workers_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/licaonfee/selina"

	"github.com/licaonfee/selina/workers"
)

func TestReaderProcess(t *testing.T) {
	errBadMarshal := errors.New("marshaler")
	errBadUnmarshal := errors.New("unmarshaler")
	badUnmarshaler := func([]byte, any) error { return errBadUnmarshal }
	badMarshaler := func(any) ([]byte, error) { return nil, errBadMarshal }
	tests := []struct {
		name    string
		data    []string
		opts    workers.TextReaderOptions
		split   bufio.SplitFunc
		want    []string
		wantErr error
	}{
		{
			name: "Success",
			data: []string{
				"Lorem ipsum dolor sit amet\n",
				"consectetur adipiscing elit\n",
				"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua"},
			opts: workers.TextReaderOptions{},
			want: []string{"Lorem ipsum dolor sit amet",
				"consectetur adipiscing elit",
				"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua"},
			wantErr: nil,
		},
		{
			name: "Success Split Words",
			data: []string{"Im a multiline text\n",
				"newlines must be ignored\n\n\n"},
			opts: workers.TextReaderOptions{
				SplitFunc: bufio.ScanWords,
			},
			want: []string{"Im", "a", "multiline", "text",
				"newlines", "must", "be", "ignored"},
			wantErr: nil,
		},
		{
			name:    "nil io.Reader",
			data:    nil,
			opts:    workers.TextReaderOptions{},
			want:    []string{},
			wantErr: workers.ErrNilTextReader,
		},
		{
			name: "read ndjson",
			data: []string{`{"name":"foo","age":25}` + "\n", `{"name":"foo","age":26}` + "\n", `{"name":"foo","age":27}` + "\n"},
			opts: workers.TextReaderOptions{
				SplitFunc:   bufio.ScanLines,
				ReadFormat:  json.Unmarshal,
				WriteFormat: json.Marshal,
			},
			want:    []string{`{"age":25,"name":"foo"}`, `{"age":26,"name":"foo"}`, `{"age":27,"name":"foo"}`},
			wantErr: nil,
		},
		{
			name: "bad marshaler",
			data: []string{`{"name":"foo","age":25}` + "\n", `{"name":"foo","age":26}` + "\n", `{"name":"foo","age":27}` + "\n"},
			opts: workers.TextReaderOptions{
				SplitFunc:   bufio.ScanLines,
				ReadFormat:  json.Unmarshal,
				WriteFormat: badMarshaler,
			},
			want:    []string{},
			wantErr: errBadMarshal,
		},
		{
			name: "bad unmarshaler",
			data: []string{`{"name":"foo","age":25}` + "\n", `{"name":"foo","age":26}` + "\n", `{"name":"foo","age":27}` + "\n"},
			opts: workers.TextReaderOptions{
				SplitFunc:   bufio.ScanLines,
				ReadFormat:  badUnmarshaler,
				WriteFormat: json.Marshal,
			},
			want:    []string{},
			wantErr: errBadUnmarshal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r io.Reader
			if tt.data != nil {
				r = strings.NewReader(strings.Join(tt.data, ""))
			}
			tt.opts.Reader = r
			w := workers.NewTextReader(tt.opts)
			input := make(chan *bytes.Buffer)
			output := make(chan *bytes.Buffer, len(tt.want))
			args := selina.ProcessArgs{Input: input, Output: output}
			err := w.Process(context.Background(), args)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Process() err = %v , want = %v ", err, tt.wantErr)
			}
			got := []string{}
			for _, b := range selina.ChannelAsSlice(output) {
				got = append(got, b.String())
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Process() got = %v , want % v", got, tt.want)
			}
		})
	}
}

func TestReaderProcessCancel(t *testing.T) {
	fileContents := []string{"fooo", "bar"}
	rd := strings.NewReader(strings.Join(fileContents, "\n"))
	tr := workers.NewTextReader(workers.TextReaderOptions{Reader: rd})
	if err := workers.ATProcessCancel(tr); err != nil {
		t.Fatal(err)
	}
}

func TestReaderProcessCloseInput(t *testing.T) {
	fileContents := []string{"fooo", "bar"}
	rd := strings.NewReader(strings.Join(fileContents, "\n"))
	tr := workers.NewTextReader(workers.TextReaderOptions{Reader: rd})
	if err := workers.ATProcessCloseInput(tr); err != nil {
		t.Fatal(err)
	}
}

func TestReaderProcessCloseOutput(t *testing.T) {
	fileContents := []string{"fooo", "bar"}
	rd := strings.NewReader(strings.Join(fileContents, "\n"))
	tr := workers.NewTextReader(workers.TextReaderOptions{Reader: rd})
	if err := workers.ATProcessCloseOutput(tr); err != nil {
		t.Fatal(err)
	}
}

func TestReaderProcessNilReader(t *testing.T) {
	opts := workers.TextReaderOptions{Reader: nil}
	tr := workers.NewTextReader(opts)
	in := make(chan *bytes.Buffer)
	out := make(chan *bytes.Buffer) // unbuffered so, process wait forever
	args := selina.ProcessArgs{Input: in, Output: out}
	err := tr.Process(context.Background(), args)
	if err != workers.ErrNilTextReader {
		t.Fatalf("Process() err = %T(%v)", err, err)
	}
}
