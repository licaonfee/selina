package input_test

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/licaonfee/selina/workers"

	"github.com/licaonfee/selina/workers/input"
)

func Test_TextReader_Process(t *testing.T) {
	fileContents := []string{
		"Lorem ipsum dolor sit amet",
		"consectetur adipiscing elit",
		"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
	}
	rd := strings.NewReader(strings.Join(fileContents, "\n"))
	opts := input.TextReaderOptions{Reader: rd}
	tr := input.NewTextReader(opts)
	input := make(chan []byte)
	output := make(chan []byte, len(fileContents))
	if err := tr.Process(context.Background(), input, output); err != nil {
		t.Fatalf("Process() err = %v", err)
	}
	got := []string{}
	for line := range output {
		got = append(got, string(line))
		if len(got) == len(fileContents) {
			break
		}
	}
	if !reflect.DeepEqual(got, fileContents) {
		t.Fatalf("Process() got = %v, want = %v", got, fileContents)
	}
}

func Test_TextReader_Process_cancel(t *testing.T) {
	fileContents := []string{"fooo", "bar"}
	rd := strings.NewReader(strings.Join(fileContents, "\n"))
	tr := input.NewTextReader(input.TextReaderOptions{Reader: rd})
	workers.ATProcessCancel(tr, t)
}

func Test_TextReader_Process_close_input(t *testing.T) {
	fileContents := []string{"fooo", "bar"}
	rd := strings.NewReader(strings.Join(fileContents, "\n"))
	tr := input.NewTextReader(input.TextReaderOptions{Reader: rd})
	workers.ATProcessCloseInput(tr, t)
}

func Test_TextReader_Process_close_output(t *testing.T) {
	fileContents := []string{"fooo", "bar"}
	rd := strings.NewReader(strings.Join(fileContents, "\n"))
	tr := input.NewTextReader(input.TextReaderOptions{Reader: rd})
	workers.ATProcessCloseOutput(tr, t)
}

func Test_TextReader_Process_NilReader(t *testing.T) {
	opts := input.TextReaderOptions{Reader: nil}
	tr := input.NewTextReader(opts)
	in := make(chan []byte)
	out := make(chan []byte) //unbuffered so, process wait forever
	err := tr.Process(context.Background(), in, out)
	if err != input.ErrNilReader {
		t.Fatalf("Process() err = %T(%v)", err, err)
	}
}
