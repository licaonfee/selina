package filesystem_test

import (
	"bufio"
	"context"
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/licaonfee/selina"
	"github.com/licaonfee/selina/workers"
	fs "github.com/licaonfee/selina/workers/filesystem"
	"github.com/spf13/afero"
)

func populateFs(files map[string]string) afero.Fs {
	m := afero.NewMemMapFs()
	for name, data := range files {
		f, _ := m.Create(name)
		f.WriteString(data)
	}
	return m
}

type nameFromBytes struct{}

func (n nameFromBytes) Filename(b []byte) string {
	return string(b)
}

func TestReaderProcess(t *testing.T) {
	tests := []struct {
		name    string
		opts    fs.ReaderOptions
		in      []string
		want    []string
		wantErr error
	}{
		{
			name: "read a file",
			opts: fs.ReaderOptions{
				Filename: fs.FilenameFunc(func() string { return "/tmp/my.file" }),
				Fs: populateFs(map[string]string{
					"/tmp/my.file": "some data\nin the file"}),
				SplitFunc: bufio.ScanLines,
			},
			in:      []string{"/tmp/my.file"},
			want:    []string{"some data", "in the file"},
			wantErr: nil,
		},
		{
			name: "read multiple files file",
			opts: fs.ReaderOptions{
				Filename: &nameFromBytes{},
				Fs: populateFs(map[string]string{
					"/tmp/01.txt":        "some data\nin the file1",
					"/tmp/otherfile.txt": "this file\nhas\n3 lines"}),
				SplitFunc: bufio.ScanLines,
			},
			in:      []string{"/tmp/01.txt", "/tmp/otherfile.txt"},
			want:    []string{"some data", "in the file1", "this file", "has", "3 lines"},
			wantErr: nil,
		},
		{
			name: "File not found",
			opts: fs.ReaderOptions{
				Filename: fs.FilenameFunc(func() string { return "/tmp/missing.file" }),
				Fs: populateFs(map[string]string{
					"/tmp/my.file": "some data\nin the file"}),
				SplitFunc: bufio.ScanLines,
			},
			in:      []string{"/tmp/my.file"},
			want:    []string{},
			wantErr: os.ErrNotExist,
		},
		{
			name: "File not found handled",
			opts: fs.ReaderOptions{
				Filename: fs.FilenameFunc(func() string { return "/tmp/missing.file" }),
				Fs: populateFs(map[string]string{
					"/tmp/my.file": "some data\nin the file"}),
				SplitFunc: bufio.ScanLines,
				Hanlder:   func(error) bool { return true },
			},
			in:      []string{"/tmp/my.file"},
			want:    []string{},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := fs.NewReader(tt.opts)
			input := selina.SliceAsChannel(tt.in, true)
			output := make(chan []byte, len(tt.want))
			args := selina.ProcessArgs{Input: input, Output: output}
			err := r.Process(context.Background(), args)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Process() err = %T , want = %v", err, tt.wantErr)
			}
			got := selina.ChannelAsSlice(output)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Process() got = %v , want = %v", got, tt.want)
			}
		})
	}
}

func TestReaderProcessCancelation(t *testing.T) {
	r := fs.NewReader(fs.ReaderOptions{})
	if err := workers.ATProcessCancel(r); err != nil {
		t.Fatal(err)
	}
}

func TestReaderCloseInput(t *testing.T) {
	r := fs.NewReader(fs.ReaderOptions{})
	if err := workers.ATProcessCloseInput(r); err != nil {
		t.Fatal(err)
	}
}

func TestReaderCloseOutput(t *testing.T) {
	r := fs.NewReader(fs.ReaderOptions{})
	if err := workers.ATProcessCloseOutput(r); err != nil {
		t.Fatal(err)
	}
}
