package output_test

import (
	"bufio"
	"bytes"
	"context"
	"reflect"
	"testing"

	"github.com/licaonfee/selina/workers"

	"github.com/licaonfee/selina"
	"github.com/licaonfee/selina/workers/output"
)

func TestTextWriter_Process(t *testing.T) {
	fileContents := []string{
		"Lorem ipsum dolor sit amet",
		"consectetur adipiscing elit",
		"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
	}
	w := &bytes.Buffer{}
	tw := output.NewTextWriter(output.TextWriterOptions{Writer: w})
	in := selina.SliceAsChannel(fileContents, true)
	out := make(chan []byte)

	if err := tw.Process(context.Background(), in, out); err != nil {
		t.Fatalf("Process() err = %v", err)
	}
	sc := bufio.NewScanner(w)
	got := []string{}
	for sc.Scan() {
		got = append(got, sc.Text())
	}
	if !reflect.DeepEqual(got, fileContents) {
		t.Fatalf("Process() got = %v , want = %v", got, fileContents)
	}
}

func Test_TextWriter_Process_NilWriter(t *testing.T) {
	opts := output.TextWriterOptions{Writer: nil}
	tr := output.NewTextWriter(opts)
	in := make(chan []byte)
	out := make(chan []byte) //unbuffered so, process wait forever
	err := tr.Process(context.Background(), in, out)
	if err != output.ErrNilWriter {
		t.Fatalf("Process() err = %T(%v)", err, err)
	}
}

func TestTextWriter_Process_close_input(t *testing.T) {
	w := &bytes.Buffer{}
	tw := output.NewTextWriter(output.TextWriterOptions{Writer: w})
	workers.ATProcessCloseInput(tw, t)
}

func TestTextWriter_Process_close_output(t *testing.T) {
	w := &bytes.Buffer{}
	tw := output.NewTextWriter(output.TextWriterOptions{Writer: w})
	workers.ATProcessCloseOutput(tw, t)
}
func TestTextWriter_Process_cancel(t *testing.T) {
	w := &bytes.Buffer{}
	tw := output.NewTextWriter(output.TextWriterOptions{Writer: w})
	workers.ATProcessCancel(tw, t)
}
