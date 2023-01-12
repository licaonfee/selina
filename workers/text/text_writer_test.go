package text_test

import (
	"bufio"
	"bytes"
	"context"
	"reflect"
	"testing"

	"github.com/licaonfee/selina/workers"
	"github.com/licaonfee/selina/workers/text"

	"github.com/licaonfee/selina"
)

func TestWriterProcess(t *testing.T) {
	fileContents := []string{
		"Lorem ipsum dolor sit amet",
		"consectetur adipiscing elit",
		"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
	}
	w := &bytes.Buffer{}
	tw := text.NewWriter(text.WriterOptions{Writer: w})
	in := selina.SliceAsChannelOfBuffer(fileContents, true)
	out := make(chan *bytes.Buffer)
	args := selina.ProcessArgs{Input: in, Output: out}
	if err := tw.Process(context.Background(), args); err != nil {
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

func TestWriterProcessNilWriter(t *testing.T) {
	opts := text.WriterOptions{Writer: nil}
	tr := text.NewWriter(opts)
	in := make(chan *bytes.Buffer)
	out := make(chan *bytes.Buffer) // unbuffered so, process wait forever
	args := selina.ProcessArgs{Input: in, Output: out}
	err := tr.Process(context.Background(), args)
	if err != text.ErrNilWriter {
		t.Fatalf("Process() err = %T(%v)", err, err)
	}
}

func TestWriterProcessCloseInput(t *testing.T) {
	w := &bytes.Buffer{}
	tw := text.NewWriter(text.WriterOptions{Writer: w})
	if err := workers.ATProcessCloseInput(tw); err != nil {
		t.Fatal(err)
	}
}

func TestWriterProcessCloseOutput(t *testing.T) {
	w := &bytes.Buffer{}
	tw := text.NewWriter(text.WriterOptions{Writer: w})
	if err := workers.ATProcessCloseOutput(tw); err != nil {
		t.Fatal(err)
	}
}
func TestWriterProcessCancel(t *testing.T) {
	w := &bytes.Buffer{}
	tw := text.NewWriter(text.WriterOptions{Writer: w})
	if err := workers.ATProcessCancel(tw); err != nil {
		t.Fatal(err)
	}
}
