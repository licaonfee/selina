package text

import (
	"bufio"
	"context"
	"errors"
	"io"

	"github.com/licaonfee/selina"
)

var _ selina.Worker = (*Reader)(nil)

// ErrNilReader is returned when a nil io.Reader interface is provided
var (
	ErrNilReader = errors.New("nil io.Reader provided to TextReader")
)

// ReaderOptions customize Reader
type ReaderOptions struct {
	// Reader from which data is readed
	Reader io.Reader
	// AutoClose if its true and Reader implements io.Closer
	// io.Reader.Close() method is called on Process finish
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
func (o ReaderOptions) Check() error {
	if o.Reader == nil {
		return ErrNilReader
	}
	return nil
}

// Reader a worker that read data from an io.Reader
type Reader struct {
	opts ReaderOptions
}

func (t *Reader) cleanup() error {
	if t.opts.Reader == nil {
		return nil
	}
	if c, ok := t.opts.Reader.(io.Closer); t.opts.AutoClose && ok {
		return c.Close()
	}
	return nil
}

// Process implements Worker interface
func (t *Reader) Process(ctx context.Context, args selina.ProcessArgs) (err error) {
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
			if err := selina.SendContext(ctx, msg, args.Output); err != nil {
				return err
			}
		}
	}
	return nil
}

// NewReader create a new Reader with given options
func NewReader(opts ReaderOptions) *Reader {
	t := Reader{opts: opts}
	return &t
}
