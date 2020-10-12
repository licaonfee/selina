package filesystem

import (
	"bufio"
	"context"
	"fmt"
	"io"

	"github.com/licaonfee/selina"
	"github.com/spf13/afero"
)

// Filenamer is used to provide an easy way to
// provide dynamic file names to reader/writer
// msg is readed from input chanel in Process
type Filenamer interface {
	Filename(msg []byte) string
}

// FilenameFunc provide a wrapper to implement a
// Filenamer with just a func()string
type FilenameFunc func() string

func (f FilenameFunc) Filename([]byte) string {
	return f()
}

type ReaderOptions struct {
	Fs        afero.Fs
	SplitFunc bufio.SplitFunc
	Filename  Filenamer
}

// Reader for every message received it call Filenamer.Filename(msg)
// to get a filename to open a given file, then read it until eof or
// context is cancelled
type Reader struct {
	opts ReaderOptions
}

func (r *Reader) Process(ctx context.Context, args selina.ProcessArgs) (err error) {
	defer close(args.Output)
	var currFile io.Closer
	defer func() {
		if currFile == nil {
			return
		}
		if e := currFile.Close(); err == nil && e != nil {
			err = e
		}
	}()
	for {
		select {
		case msg, ok := <-args.Input:
			if !ok {
				return nil
			}
			fname := r.opts.Filename.Filename(msg)
			file, err := r.opts.Fs.Open(fname)
			if err != nil {
				return fmt.Errorf("Process was unable to open file from fs %w", err)
			}
			currFile = file
			sc := bufio.NewScanner(file)
			sc.Split(r.opts.SplitFunc)
			if err := readFile(ctx, sc, args.Output); err != nil {
				return err
			}
			currFile = nil
			if err := file.Close(); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}

	}
}

func readFile(ctx context.Context, sc *bufio.Scanner, out chan<- []byte) error {
	for sc.Scan() {
		select {
		case out <- []byte(sc.Text()):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func NewReader(opts ReaderOptions) *Reader {
	return &Reader{opts: opts}
}
