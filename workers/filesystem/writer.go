package filesystem

import (
	"context"
	"io"
	"os"

	"github.com/licaonfee/selina"
	"github.com/spf13/afero"
)

type WriterOptions struct {
	Filename   Filenamer
	Fs         afero.Fs
	AddNewLine bool
	BufferSize int
	Mode       os.FileMode
	Hanlder    selina.ErrorHandler
}

type Writer struct {
	opts WriterOptions
}

func (w Writer) Process(ctx context.Context, args selina.ProcessArgs) (err error) {
	defer close(args.Output)
	var currFname string
	var currFile io.WriteCloser
	if w.opts.Mode == 0 {
		w.opts.Mode = 0600
	}
	defer func() {
		if currFile == nil {
			return
		}
		if e := currFile.Close(); err == nil && e != nil {
			err = e
		}
	}()
	errHandler := selina.DefaultErrorHanler
	if w.opts.Hanlder != nil {
		errHandler = w.opts.Hanlder
	}
	for {
		select {
		case msg, ok := <-args.Input:
			if !ok {
				return nil
			}
			fname := w.opts.Filename.Filename(msg)
			if fname != currFname {
				if currFile != nil {
					currFile.Close()
				}
				f, err := w.opts.Fs.Create(fname)
				switch {
				case err == nil:
				case errHandler(err):
					continue
				default:
					return err
				}
				err = w.opts.Fs.Chmod(fname, w.opts.Mode)
				switch {
				case err == nil:
				case errHandler(err):
					continue
				default:
					return err
				}
				currFile = f
				currFname = fname
			}
			if _, err := currFile.Write(msg); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func NewWriter(opts WriterOptions) *Writer {
	return &Writer{opts: opts}
}
