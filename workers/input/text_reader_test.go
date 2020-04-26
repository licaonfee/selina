package input_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/licaonfee/selina/workers"

	"github.com/licaonfee/selina/workers/input"
)

func prepareFile(contents []string) (f *os.File, cleanup func()) {
	f, err := ioutil.TempFile("", "TextReader_Process")
	if err != nil {
		panic(fmt.Sprintf("prepareFile: %v", err))
	}
	for _, line := range contents {
		if _, err := fmt.Fprintln(f, line); err != nil {
			panic(fmt.Sprintf("prepareFile: %v", err))
		}
	}
	f.Close()
	cleanup = func() {
		os.Remove(f.Name())
	}
	return f, cleanup
}

func Test_TextReader_Process(t *testing.T) {
	fileContents := []string{
		"Lorem ipsum dolor sit amet",
		"consectetur adipiscing elit",
		"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
	}
	f, cleanup := prepareFile(fileContents)
	defer cleanup()
	opts := input.TextReaderOptions{Filename: f.Name()}
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
	f, cleanup := prepareFile(fileContents)
	defer cleanup()
	tr := input.NewTextReader(input.TextReaderOptions{Filename: f.Name()})
	workers.ATProcessCancel(tr, t)
}

func Test_TextReader_Process_close_input(t *testing.T) {
	fileContents := []string{"fooo", "bar"}
	f, cleanup := prepareFile(fileContents)
	defer cleanup()
	tr := input.NewTextReader(input.TextReaderOptions{Filename: f.Name()})
	workers.ATProcessCloseInput(tr, t)
}

func Test_TextReader_Process_close_output(t *testing.T) {
	fileContents := []string{"foo", "bar"}
	f, cleanup := prepareFile(fileContents)
	defer cleanup()
	tr := input.NewTextReader(input.TextReaderOptions{Filename: f.Name()})
	workers.ATProcessCloseOutput(tr, t)
}

func Test_TextReader_Process_MissingFile(t *testing.T) {
	opts := input.TextReaderOptions{Filename: "missing.txt"}
	tr := input.NewTextReader(opts)
	input := make(chan []byte)
	output := make(chan []byte) //unbuffered so, process wait forever
	err := tr.Process(context.Background(), input, output)
	if errAssert, ok := err.(*os.PathError); !ok {
		t.Fatalf("Process() err = %T(%v)", errAssert, errAssert)
	}
}
