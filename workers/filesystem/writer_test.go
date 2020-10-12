package filesystem_test

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/licaonfee/selina"
	"github.com/licaonfee/selina/workers"
	fs "github.com/licaonfee/selina/workers/filesystem"
	"github.com/spf13/afero"
)

func compareFs(a, b afero.Fs) bool {
	finfo, err := afero.ReadDir(a, "/tmp/")
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	for _, info := range finfo {
		name := filepath.Join("/tmp/", info.Name())
		af, err := a.Open(name)
		if err != nil {
			fmt.Printf("%v\n", err)
			return false
		}
		bf, err := b.Open(name)
		if err != nil {
			fmt.Printf("%v\n", err)
			return false
		}
		ba, _ := ioutil.ReadAll(af)
		bb, _ := ioutil.ReadAll(bf)
		if !reflect.DeepEqual(ba, bb) {
			fmt.Printf("%s != %s\n", ba, bb)
			return false
		}
	}
	return true
}

func TestWriterProcess(t *testing.T) {
	tests := []struct {
		name    string
		opts    fs.WriterOptions
		in      []string
		want    afero.Fs
		wantErr error
	}{
		{
			name: "write a file",
			opts: fs.WriterOptions{
				Filename: fs.FilenameFunc(func() string { return "/tmp/my.file" }),
				Fs:       afero.NewMemMapFs(),
			},
			in: []string{"write\n", "this\n", "data"},
			want: populateFs(map[string]string{
				"/tmp/my.file": "write\nthis\ndata"}),
			wantErr: nil,
		},
		{
			name: "write multiple files file",
			opts: fs.WriterOptions{
				Filename: &nameFromBytes{},
				Fs:       afero.NewMemMapFs(),
			},
			in: []string{"/tmp/01.txt", "/tmp/otherfile.txt"},
			want: populateFs(map[string]string{
				"/tmp/01.txt":        "/tmp/01.txt",
				"/tmp/otherfile.txt": "/tmp/otherfile.txt"}),
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := fs.NewWriter(tt.opts)
			input := selina.SliceAsChannel(tt.in, true)
			output := make(chan []byte)
			args := selina.ProcessArgs{Input: input, Output: output}
			err := r.Process(context.Background(), args)
			if !errors.Is(err, tt.wantErr) && !errors.As(err, &tt.wantErr) {
				t.Errorf("Process() err = %T , want = %v", err, tt.wantErr)
			}

			if !compareFs(tt.want, tt.opts.Fs) {
				t.Errorf("Process() Fs mismatch")
			}
		})
	}
}

func TestWriterProcessCancelation(t *testing.T) {
	r := fs.NewWriter(fs.WriterOptions{})
	if err := workers.ATProcessCancel(r); err != nil {
		t.Fatal(err)
	}
}

func TestWriterCloseInput(t *testing.T) {
	r := fs.NewWriter(fs.WriterOptions{})
	if err := workers.ATProcessCloseInput(r); err != nil {
		t.Fatal(err)
	}
}

func TestWriterCloseOutput(t *testing.T) {
	r := fs.NewWriter(fs.WriterOptions{})
	if err := workers.ATProcessCloseOutput(r); err != nil {
		t.Fatal(err)
	}
}
