package text_test

import (
	"bufio"
	"context"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/licaonfee/selina"

	"github.com/licaonfee/selina/workers"
	"github.com/licaonfee/selina/workers/text"
)

func TestReaderProcess(t *testing.T) {
	tests := []struct {
		name    string
		data    []string
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
			split: nil,
			want: []string{"Lorem ipsum dolor sit amet",
				"consectetur adipiscing elit",
				"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua"},
			wantErr: nil,
		},
		{
			name: "Success Split Words",
			data: []string{"Im a multiline text\n",
				"newlines must be ignored\n\n\n"},
			split: bufio.ScanWords,
			want: []string{"Im", "a", "multiline", "text",
				"newlines", "must", "be", "ignored"},
			wantErr: nil,
		},
		{
			name:    "nil io.Reader",
			data:    nil,
			split:   nil,
			want:    []string{},
			wantErr: text.ErrNilReader,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r io.Reader
			if tt.data != nil {
				r = strings.NewReader(strings.Join(tt.data, ""))
			}
			opts := text.ReaderOptions{Reader: r, SplitFunc: tt.split}
			w := text.NewReader(opts)
			input := make(chan []byte)
			output := make(chan []byte, len(tt.want))
			args := selina.ProcessArgs{Input: input, Output: output}
			err := w.Process(context.Background(), args)
			if err != tt.wantErr && !errors.As(err, &tt.wantErr) {
				t.Errorf("Process() err = %v , want = %v ", err, tt.wantErr)
			}
			got := selina.ChannelAsSlice(output)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Process() got = %v , want % v", got, tt.want)
			}
		})
	}
}

func TestReaderProcessCancel(t *testing.T) {
	fileContents := []string{"fooo", "bar"}
	rd := strings.NewReader(strings.Join(fileContents, "\n"))
	tr := text.NewReader(text.ReaderOptions{Reader: rd})
	if err := workers.ATProcessCancel(tr); err != nil {
		t.Fatal(err)
	}
}

func TestReaderProcessCloseInput(t *testing.T) {
	fileContents := []string{"fooo", "bar"}
	rd := strings.NewReader(strings.Join(fileContents, "\n"))
	tr := text.NewReader(text.ReaderOptions{Reader: rd})
	if err := workers.ATProcessCloseInput(tr); err != nil {
		t.Fatal(err)
	}
}

func TestReaderProcessCloseOutput(t *testing.T) {
	fileContents := []string{"fooo", "bar"}
	rd := strings.NewReader(strings.Join(fileContents, "\n"))
	tr := text.NewReader(text.ReaderOptions{Reader: rd})
	if err := workers.ATProcessCloseOutput(tr); err != nil {
		t.Fatal(err)
	}
}

func TestReaderProcessNilReader(t *testing.T) {
	opts := text.ReaderOptions{Reader: nil}
	tr := text.NewReader(opts)
	in := make(chan []byte)
	out := make(chan []byte) //unbuffered so, process wait forever
	args := selina.ProcessArgs{Input: in, Output: out}
	err := tr.Process(context.Background(), args)
	if err != text.ErrNilReader {
		t.Fatalf("Process() err = %T(%v)", err, err)
	}
}
