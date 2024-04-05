package workers

import (
	"bufio"
	"context"
	"errors"
	"io"

	"github.com/licaonfee/selina"
)

var _ selina.Worker = (*TextReader)(nil)

// ErrNilTextReader is returned when a nil io.TextReader interface is provided
var (
	ErrNilTextReader = errors.New("nil io.TextReader provided to TextReader")
)

// TextReaderOptions customize TextReader
type TextReaderOptions struct {
	// Reader from which data is readed
	Reader io.Reader
	// AutoClose if its true and TextReader implements io.Closer
	// io.TextReader.Close() method is called on Process finish
	AutoClose bool
	// Default is ScanLines
	SplitFunc bufio.SplitFunc
	// ReadFormat process every data point with this function
	// default is nil , raw message is passed to WriteFormat
	ReadFormat selina.Unmarshaler
	// WriteFormat by default is json.Marshal
	WriteFormat selina.Marshaler
}

// Check if a combination of options is valid
func (o TextReaderOptions) Check() error {
	if o.Reader == nil {
		return ErrNilTextReader
	}
	return nil
}

// TextReader a worker that read data from an io.TextReader
type TextReader struct {
	opts TextReaderOptions
}

func (t *TextReader) cleanup() error {
	if t.opts.Reader == nil {
		return nil
	}
	if c, ok := t.opts.Reader.(io.Closer); t.opts.AutoClose && ok {
		return c.Close()
	}
	return nil
}

// Process implements Worker interface
func (t *TextReader) Process(ctx context.Context, args selina.ProcessArgs) (err error) {
	defer func() {
		close(args.Output)
		cerr := t.cleanup()
		if err == nil { // if an error occurred not override it
			err = cerr
		}
	}()
	if err := t.opts.Check(); err != nil {
		return err
	}
	sc := bufio.NewScanner(t.opts.Reader)
	if t.opts.SplitFunc != nil {
		sc.Split(t.opts.SplitFunc)
	}
	wf := selina.DefaultMarshaler
	if t.opts.WriteFormat != nil {
		wf = t.opts.WriteFormat
	}
	for sc.Scan() {
		select {
		case _, ok := <-args.Input:
			if !ok {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg := []byte(sc.Text())
			if t.opts.ReadFormat != nil {
				data := new(interface{})
				if err := t.opts.ReadFormat(msg, data); err != nil {
					return err
				}
				msg, err = wf(data)
				if err != nil {
					return err
				}
			}
			b := selina.GetBuffer()
			b.Write(msg)
			if err := selina.SendContext(ctx, b, args.Output); err != nil {
				return err
			}
		}
	}
	return nil
}

// NewTextReader create a new TextReader with given options
func NewTextReader(opts TextReaderOptions) *TextReader {
	t := TextReader{opts: opts}
	return &t
}
