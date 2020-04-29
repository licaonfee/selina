package text

import (
	"bufio"
	"context"
	"errors"
	"io"

	"github.com/licaonfee/selina"
)

var _ selina.Worker = (*Writer)(nil)

type WriterOptions struct {
	Writer    io.Writer
	AutoClose bool
}

type Writer struct {
	opts WriterOptions
}

func (t *Writer) cleanup() error {
	if t.opts.Writer == nil {
		return nil
	}
	if c, ok := t.opts.Writer.(io.Closer); t.opts.AutoClose && ok {
		return c.Close()
	}
	return nil
}

var ErrNilWriter = errors.New("nil io.Writer provided to TextWriter")

func (t *Writer) Process(ctx context.Context, in <-chan []byte, out chan<- []byte) (err error) {
	defer func() {
		close(out)
		cerr := t.cleanup()
		if err == nil { //if an error occurred not override it
			err = cerr
		}
	}()
	if t.opts.Writer == nil {
		return ErrNilWriter
	}
	w := bufio.NewWriter(t.opts.Writer)
	defer func() {
		if errFlush := w.Flush(); errFlush != nil {
			err = errFlush
		}
	}()
	newLine := []byte("\n")
	for {
		select {
		case msg, ok := <-in:
			if !ok {
				return nil
			}
			_, err = w.Write(msg)
			if err != nil {
				return
			}
			_, err = w.Write(newLine)
			if err != nil {
				return
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func NewTextWriter(opts WriterOptions) *Writer {
	w := &Writer{opts: opts}
	return w
}
