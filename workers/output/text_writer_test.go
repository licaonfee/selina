package output_test

import (
	"bufio"
	"context"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/licaonfee/selina/workers"

	"github.com/licaonfee/selina"
	"github.com/licaonfee/selina/workers/output"
)

func prepareFileName() string {
	file, err := ioutil.TempFile("", "Text_Writer")
	if err != nil {
		panic(err)
	}
	os.Remove(file.Name())
	return file.Name()
}
func TestTextWriter_Process(t *testing.T) {
	fileContents := []string{
		"Lorem ipsum dolor sit amet",
		"consectetur adipiscing elit",
		"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
	}
	filename := prepareFileName()
	opts := output.TextWriterOptions{Filename: filename}
	tw := output.NewTextWriter(opts)
	input := selina.SliceAsChannel(fileContents, true)
	output := make(chan []byte)

	if err := tw.Process(context.Background(), input, output); err != nil {
		t.Fatalf("Process() err = %v", err)
	}
	f, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Process() err = %v", err)
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	got := []string{}
	for sc.Scan() {
		got = append(got, sc.Text())
	}
	if !reflect.DeepEqual(got, fileContents) {
		t.Fatalf("Process() got = %v , want = %v", got, fileContents)
	}
}

func TestTextWriter_Process_close_input(t *testing.T) {
	filename := prepareFileName()
	opts := output.TextWriterOptions{Filename: filename}
	tw := output.NewTextWriter(opts)
	workers.ATProcessCloseInput(tw, t)
}

func TestTextWriter_Process_close_output(t *testing.T) {
	filename := prepareFileName()
	opts := output.TextWriterOptions{Filename: filename}
	tw := output.NewTextWriter(opts)
	workers.ATProcessCloseOutput(tw, t)
}
func TestTextWriter_Process_cancel(t *testing.T) {
	filename := prepareFileName()
	opts := output.TextWriterOptions{Filename: filename}
	tw := output.NewTextWriter(opts)
	workers.ATProcessCancel(tw, t)
}
