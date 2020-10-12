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
	for {
		select {
		case msg, ok := <-args.Input:
			if !ok {
				return nil
			}
			fname := w.opts.Filename.Filename(msg)
			if fname != currFname {
				currFname = fname
				if currFile != nil {
					currFile.Close()
				}

				f, err := w.opts.Fs.Create(fname)
				if err != nil {
					return err
				}
				currFile = f
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
