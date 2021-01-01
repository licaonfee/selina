package text

import (
	"bufio"
	"context"
	"errors"
	"io"

	"github.com/licaonfee/selina"
)

var _ selina.Worker = (*Writer)(nil)

//WriterOptions customize Writer
type WriterOptions struct {
	//Writer io.Writer where data will be written
	Writer io.Writer
	//AutoClose when true and Writer implements io.Closer
	// io.Closer.Close() method will be called on finalization
	AutoClose   bool
	SkipNewLine bool
	BufferSize  int
	Codec       selina.Marshaler
}

//Check if a combination of options is valid
func (o WriterOptions) Check() error {
	if o.Writer == nil {
		return ErrNilWriter
	}
	return nil
}

//Writer a Worker that write data to a given io.Writer in text format
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

//ErrNilWriter returned when a nil io.Writer is provided
var ErrNilWriter = errors.New("nil io.Writer provided to TextWriter")

//Process implements Worker interface
func (t *Writer) Process(ctx context.Context, args selina.ProcessArgs) (err error) {
	defer func() {
		close(args.Output)
		cerr := t.cleanup()
		if err == nil { //if an error occurred not override it
			err = cerr
		}
	}()
	if err := t.opts.Check(); err != nil {
		return err
	}
	w := bufio.NewWriterSize(t.opts.Writer, t.opts.BufferSize)
	defer func() {
		if errFlush := w.Flush(); errFlush != nil {
			err = errFlush
		}
	}()
	newLine := []byte("\n")
	for {
		select {
		case msg, ok := <-args.Input:
			if !ok {
				return nil
			}
			if t.opts.Codec != nil {
				data, err := t.opts.Codec(msg)
				if err != nil {
					return err
				}
				msg = data
			}
			_, err = w.Write(msg)
			if err != nil {
				return
			}
			if !t.opts.SkipNewLine {
				_, err = w.Write(newLine)
				if err != nil {
					return
				}
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

//NewWriter create a new Writer with given options
func NewWriter(opts WriterOptions) *Writer {
	w := &Writer{opts: opts}
	return w
}
