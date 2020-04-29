package text_test

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/licaonfee/selina/workers"
	"github.com/licaonfee/selina/workers/text"
)

func TestReaderProcess(t *testing.T) {
	fileContents := []string{
		"Lorem ipsum dolor sit amet",
		"consectetur adipiscing elit",
		"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
	}
	rd := strings.NewReader(strings.Join(fileContents, "\n"))
	opts := text.ReaderOptions{Reader: rd}
	tr := text.NewReader(opts)
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

func TestReaderProcessCancel(t *testing.T) {
	fileContents := []string{"fooo", "bar"}
	rd := strings.NewReader(strings.Join(fileContents, "\n"))
	tr := text.NewReader(text.ReaderOptions{Reader: rd})
	workers.ATProcessCancel(tr, t)
}

func TestReaderProcessCloseInput(t *testing.T) {
	fileContents := []string{"fooo", "bar"}
	rd := strings.NewReader(strings.Join(fileContents, "\n"))
	tr := text.NewReader(text.ReaderOptions{Reader: rd})
	workers.ATProcessCloseInput(tr, t)
}

func TestReaderProcessCloseOutput(t *testing.T) {
	fileContents := []string{"fooo", "bar"}
	rd := strings.NewReader(strings.Join(fileContents, "\n"))
	tr := text.NewReader(text.ReaderOptions{Reader: rd})
	workers.ATProcessCloseOutput(tr, t)
}

func TestReaderProcessNilReader(t *testing.T) {
	opts := text.ReaderOptions{Reader: nil}
	tr := text.NewReader(opts)
	in := make(chan []byte)
	out := make(chan []byte) //unbuffered so, process wait forever
	err := tr.Process(context.Background(), in, out)
	if err != text.ErrNilReader {
		t.Fatalf("Process() err = %T(%v)", err, err)
	}
}
