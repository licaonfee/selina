// Package filesystem utilities for read and write files
package filesystem

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/licaonfee/selina"
	"github.com/spf13/afero"
)

// Filenamer is used to provide an easy way to
// generate dynamic files name to reader/writer
// msg is readed from input chanel in Process
type Filenamer interface {
	Filename(msg []byte) string
}

// FilenameFunc provide a wrapper to implement a
// Filenamer with just a func()string
type FilenameFunc func() string

// Filename return a filename to be opened by worker
func (f FilenameFunc) Filename([]byte) string {
	return f()
}

// ReaderOptions configuration for Reader worker
type ReaderOptions struct {
	Fs        afero.Fs
	SplitFunc bufio.SplitFunc
	Filename  Filenamer
	Handler   selina.ErrorHandler
}

// Reader for every message received it call Filenamer.Filename(msg)
// to get a filename to open a given file, then read it until eof or
// context is cancelled
type Reader struct {
	opts ReaderOptions
}

// Process see selina.Worker
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
	errHandler := selina.DefaultErrorHandler
	if r.opts.Handler != nil {
		errHandler = r.opts.Handler
	}
	for {
		select {
		case msg, ok := <-args.Input:
			if !ok {
				return nil
			}
			fname := r.opts.Filename.Filename(msg.Bytes())
			selina.FreeBuffer(msg)
			file, err := r.opts.Fs.Open(fname)
			switch {
			case err == nil:
			case errHandler(err):
				continue
			default:
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

func readFile(ctx context.Context, sc *bufio.Scanner, out chan<- *bytes.Buffer) error {
	for sc.Scan() {
		msg := selina.GetBuffer()
		msg.Write(sc.Bytes())
		select {
		case out <- msg:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

// NewReader create a new reader with goven options
func NewReader(opts ReaderOptions) *Reader {
	return &Reader{opts: opts}
}
